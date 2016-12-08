package enmime

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
)

// Part is the primary structure enmine clients will interact with.  Each Part represents a
// node in the MIME multipart tree.  The Content-Type, Disposition and File Name are parsed out of
// the header for easier access.
type Part struct {
	Header      textproto.MIMEHeader // Header for this Part
	Parent      *Part                // Parent of this part (can be nil)
	FirstChild  *Part                // FirstChild is the top most child of this part
	NextSibling *Part                // NextSibling of this part
	ContentType string               // ContentType header without parameters
	Disposition string               // Content-Disposition header without parameters
	FileName    string               // The file-name from disposition or type header
	Charset     string               // The content charset encoding label
	Errors      []Error              // Errors encountered while parsing this part

	boundary      string    // Boundary marker used within this part
	rawReader     io.Reader // The raw Part content, no decoding or charset conversion
	decodedReader io.Reader // The content decoded from quoted-printable or base64
	utf8Reader    io.Reader // The decoded content converted to UTF-8
}

// NewPart creates a new Part object.  It does not update the parents FirstChild attribute.
func NewPart(parent *Part, contentType string) *Part {
	return &Part{Parent: parent, ContentType: contentType}
}

// Read returns the decoded & UTF-8 converted content; implements io.Reader.
func (p *Part) Read(b []byte) (n int, err error) {
	if p.utf8Reader == nil {
		return 0, io.EOF
	}
	return p.utf8Reader.Read(b)
}

// setupContentHeaders uses Content-Type media params and Content-Disposition headers to populate
// the disposition, filename, and charset fields.
func (p *Part) setupContentHeaders(mediaParams map[string]string) {
	// Determine content disposition, filename, character set
	disposition, dparams, err := mime.ParseMediaType(p.Header.Get(hnContentDisposition))
	if err == nil {
		// Disposition is optional
		p.Disposition = disposition
		p.FileName = decodeHeader(dparams[hpFilename])
	}
	if p.FileName == "" && mediaParams[hpName] != "" {
		p.FileName = decodeHeader(mediaParams[hpName])
	}
	if p.FileName == "" && mediaParams[hpFile] != "" {
		p.FileName = decodeHeader(mediaParams[hpFile])
	}
	if p.Charset == "" {
		p.Charset = mediaParams[hpCharset]
	}
}

// buildContentReaders sets up the decodedReader and utf8Reader based on the Part headers.  If no
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
	encoding := p.Header.Get(hnContentEncoding)
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
		if p.Charset != "" {
			if reader, err := newCharsetReader(p.Charset, contentReader); err == nil {
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
	root := &Part{}

	// Read header
	header, err := readHeader(br, root)
	if err != nil {
		return nil, err
	}
	root.Header = header

	// Content-Type
	contentType := header.Get(hnContentType)
	if contentType == "" {
		root.addWarning(
			errorMissingContentType,
			"MIME parts should have a Content-Type header")
	}
	mediatype, params, err := mime.ParseMediaType(header.Get(hnContentType))
	if contentType != "" && err != nil {
		return nil, err
	}
	root.ContentType = mediatype
	root.Charset = params[hpCharset]

	if strings.HasPrefix(mediatype, ctMultipartPrefix) {
		// Content is multipart, parse it
		boundary := params[hpBoundary]
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
func parseParts(parent *Part, reader *bufio.Reader, boundary string) error {
	var prevSibling *Part

	// Loop over MIME parts
	br := newBoundaryReader(reader, boundary)
	for {
		next, err := br.Next()
		if err != nil && err != io.EOF {
			return err
		}
		if !next {
			break
		}
		p := &Part{Parent: parent}
		bbr := bufio.NewReader(br)
		header, err := readHeader(bbr, p)
		p.Header = header
		if err == errEmptyHeaderBlock {
			// Empty header probably means the part didn't use the correct trailing "--" syntax to
			// close its boundary.
			if next, err = br.Next(); err != nil {
				if err == io.EOF || strings.HasSuffix(err.Error(), "EOF") {
					// There are no more Parts, but the error belongs to a sibling or parent,
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
				}
				return fmt.Errorf("Error at boundary %v: %v", boundary, err)
			}
		} else if err != nil {
			return err
		}

		ctype := header.Get(hnContentType)
		if ctype == "" {
			p.addWarning(
				errorMissingContentType,
				"MIME parts should have a Content-Type header")
		} else {
			// Parse Content-Type header
			mtype, mparams, err := mime.ParseMediaType(ctype)
			if err != nil {
				return err
			}
			p.ContentType = mtype

			// Set disposition, filename, charset if available
			p.setupContentHeaders(mparams)
			p.boundary = mparams[hpBoundary]
		}

		// Insert this Part into the MIME tree
		if prevSibling != nil {
			prevSibling.NextSibling = p
		} else {
			parent.FirstChild = p
		}
		prevSibling = p

		if p.boundary != "" {
			// Content is another multipart
			err = parseParts(p, bbr, p.boundary)
			if err != nil {
				return err
			}
		} else {
			// Content is text or data: build content reader pipeline
			if err := p.buildContentReaders(bbr); err != nil {
				return err
			}
		}
	}

	return nil
}
