package enmime

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
)

// Part is the primary structure enmine clients will interact with.  Each Part represents a
// node in the MIME multipart tree.  The Content-Type, Disposition and File Name are parsed out of
// the header for easier access.
type Part struct {
	Header        textproto.MIMEHeader // Header for this Part
	parent        *Part
	firstChild    *Part
	nextSibling   *Part
	contentType   string
	disposition   string
	fileName      string
	charset       string
	errors        []Error   // Errors encountered while parsing this part
	rawReader     io.Reader // The raw Part content, no decoding or charset conversion
	decodedReader io.Reader // The content decoded from quoted-printable or base64
	utf8Reader    io.Reader // The decoded content converted to UTF-8
}

// NewPart creates a new Part object.  It does not update the parents FirstChild attribute.
func NewPart(parent *Part, contentType string) *Part {
	return &Part{parent: parent, contentType: contentType}
}

// Parent of this part (can be nil)
func (p *Part) Parent() *Part {
	return p.parent
}

// SetParent sets the parent of this part
func (p *Part) SetParent(parent *Part) {
	p.parent = parent
}

// FirstChild is the top most child of this part
func (p *Part) FirstChild() *Part {
	return p.firstChild
}

// SetFirstChild sets the first (top most) child of this part
func (p *Part) SetFirstChild(child *Part) {
	p.firstChild = child
}

// NextSibling of this part
func (p *Part) NextSibling() *Part {
	return p.nextSibling
}

// SetNextSibling sets the next sibling (shares parent) of this part
func (p *Part) SetNextSibling(sibling *Part) {
	p.nextSibling = sibling
}

// ContentType header without parameters
func (p *Part) ContentType() string {
	return p.contentType
}

// SetContentType sets the Content-Type.
//
// Example: "image/jpg" or "application/octet-stream"
func (p *Part) SetContentType(contentType string) {
	p.contentType = contentType
}

// Disposition returns the Content-Disposition header without parameters
func (p *Part) Disposition() string {
	return p.disposition
}

// SetDisposition sets the Content-Disposition.
//
// Example: "attachment" or "inline"
func (p *Part) SetDisposition(disposition string) {
	p.disposition = disposition
}

// FileName returns the file name from disposition or type header
func (p *Part) FileName() string {
	return p.fileName
}

// SetFileName sets the parts file name.
func (p *Part) SetFileName(fileName string) {
	p.fileName = fileName
}

// Charset returns the content charset encoding label
func (p *Part) Charset() string {
	return p.charset
}

// SetCharset sets the charset encoding label of this content, see charsets.go
// for examples, but you probably want "utf-8"
func (p *Part) SetCharset(charset string) {
	p.charset = charset
}

// Errors returns a slice of Errors encountered while parsing this Part
func (p *Part) Errors() []Error {
	return p.errors
}

// Read implements io.Reader
func (p *Part) Read(b []byte) (n int, err error) {
	if p.utf8Reader == nil {
		return 0, io.EOF
	}
	return p.utf8Reader.Read(b)
}

// buildContentReaders sets up the decodedReader and ut8Reader based on the Part headers.  If no
// translation is required at a particular stage, the reader will be the same as its predecessor.
// If the content encoding type is not recognized, no effort will be made to do character set
// conversion.
func (p *Part) buildContentReaders(r io.Reader) error {
	// Read raw content into buffer
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	var contentReader io.Reader = buf
	valid := true

	// Raw content reader
	p.rawReader = contentReader

	// Build content decoding reader
	encoding := p.Header.Get("Content-Transfer-Encoding")
	switch strings.ToLower(encoding) {
	case "quoted-printable":
		contentReader = quotedprintable.NewReader(contentReader)
	case "base64":
		contentReader = newBase64Cleaner(contentReader)
		contentReader = base64.NewDecoder(base64.StdEncoding, contentReader)
	case "8bit", "7bit", "binary", "":
		// No decoding required
	default:
		// Unknown encoding
		valid = false
		p.addWarning(
			errorContentEncoding,
			"Unrecognized Content-Transfer-Encoding type %q",
			encoding)
	}
	p.decodedReader = contentReader

	if valid {
		// decodedReader is good; build character set conversion reader
		if p.Charset() != "" {
			if reader, err := newCharsetReader(p.Charset(), contentReader); err == nil {
				contentReader = reader
			} else {
				// Failed to get a conversion reader
				p.addWarning(errorCharsetConversion, err.Error())
			}
		}
	}
	p.utf8Reader = contentReader
	return nil
}

// ReadParts reads a MIME document from the provided reader and parses it into tree of Part objects.
func ReadParts(r io.Reader) (*Part, error) {
	br := bufio.NewReader(r)

	// Read header
	tr := textproto.NewReader(br)
	header, err := tr.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	root := &Part{Header: header}

	// Content-Type
	contentType := header.Get("Content-Type")
	if contentType == "" {
		root.addWarning(
			errorMissingContentType,
			"MIME parts should have a Content-Type header")
	}
	mediatype, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if contentType != "" && err != nil {
		return nil, err
	}
	root.contentType = mediatype
	root.charset = params["charset"]

	if strings.HasPrefix(mediatype, "multipart/") {
		// Content is multipart, parse it
		boundary := params["boundary"]
		err = parseParts(root, br, boundary)
		if err != nil {
			return nil, err
		}
	} else {
		// Content is text or data, build content reader pipeline
		if err := root.buildContentReaders(br); err != nil {
			return nil, err
		}
	}

	return root, nil
}

// parseParts recursively parses a mime multipart document.
func parseParts(parent *Part, reader io.Reader, boundary string) error {
	var prevSibling *Part

	// Loop over MIME parts
	mr := multipart.NewReader(reader, boundary)
	for {
		// mrp is golang's built in mime-part
		mrp, err := mr.NextPart()
		if err != nil {
			if err == io.EOF {
				// This is a clean end-of-message signal
				break
			}
			return err
		}
		if len(mrp.Header) == 0 {
			// Empty header probably means the part didn't use the correct trailing "--" syntax to
			// close its boundary.  We will let this slide if this this the last MIME part.
			if _, err = mr.NextPart(); err != nil {
				if err == io.EOF || strings.HasSuffix(err.Error(), "EOF") {
					// There are no more MIME parts, but the error belongs to our sibling or parent,
					// because this Part doesn't actually exist.
					owner := parent
					if prevSibling != nil {
						owner = prevSibling
					}
					owner.addWarning(
						errorMissingBoundary,
						"Boundary %q was not closed correctly",
						boundary)
					break
				} else {
					return fmt.Errorf("Error at boundary %v: %v", boundary, err)
				}
			}

			return fmt.Errorf("Empty header at boundary %v", boundary)
		}
		ctype := mrp.Header.Get("Content-Type")
		if ctype == "" {
			return fmt.Errorf("Missing Content-Type at boundary %v", boundary)
		}
		mediatype, mparams, err := mime.ParseMediaType(ctype)
		if err != nil {
			return err
		}

		// Insert ourselves into tree, p is enmime's MIME part
		p := NewPart(parent, mediatype)
		p.Header = mrp.Header
		if prevSibling != nil {
			prevSibling.SetNextSibling(p)
		} else {
			parent.SetFirstChild(p)
		}
		prevSibling = p

		// Determine content disposition, filename, character set
		disposition, dparams, err := mime.ParseMediaType(mrp.Header.Get("Content-Disposition"))
		if err == nil {
			// Disposition is optional
			p.SetDisposition(disposition)
			p.SetFileName(decodeHeader(dparams["filename"]))
		}
		if p.FileName() == "" && mparams["name"] != "" {
			p.SetFileName(decodeHeader(mparams["name"]))
		}
		if p.FileName() == "" && mparams["file"] != "" {
			p.SetFileName(decodeHeader(mparams["file"]))
		}
		if p.Charset() == "" {
			p.SetCharset(mparams["charset"])
		}

		boundary := mparams["boundary"]
		if boundary != "" {
			// Content is another multipart
			err = parseParts(p, mrp, boundary)
			if err != nil {
				return err
			}
		} else {
			// Content is text or data, build content reader pipeline
			if err := p.buildContentReaders(mrp); err != nil {
				return err
			}
		}
	}

	return nil
}
