package enmime

import (
	"fmt"
	"github.com/jaytaylor/html2text"
	"io"
	"io/ioutil"
	"mime"
	"net/mail"
	"net/textproto"
	"strings"
)

// AddressHeaders is the set of SMTP headers that contain email addresses, used by
// Envelope.AddressList().  Key characters must be all lowercase.
var AddressHeaders = map[string]bool{
	"bcc":          true,
	"cc":           true,
	"delivered-to": true,
	"from":         true,
	"reply-to":     true,
	"to":           true,
}

// Envelope is a simplified wrapper for MIME email messages.
type Envelope struct {
	Text           string                // The plain text portion of the message
	HTML           string                // The HTML portion of the message
	IsTextFromHTML bool                  // Plain text was empty; down-converted HTML
	Root           *Part                 // The top-level Part
	Attachments    []*Part               // All parts having a Content-Disposition of attachment
	Inlines        []*Part               // All parts having a Content-Disposition of inline
	OtherParts     []*Part               // All parts not in Attachments and Inlines
	Errors         []*Error              // Errors encountered while parsing
	header         *textproto.MIMEHeader // Header from original message
}

// GetHeader processes the specified header for RFC 2047 encoded words and returns the result as a
// UTF-8 string
func (e *Envelope) GetHeader(name string) string {
	if e.header == nil {
		return ""
	}
	return decodeHeader(e.header.Get(name))
}

// AddressList returns a mail.Address slice with RFC 2047 encoded names converted to UTF-8
func (e *Envelope) AddressList(key string) ([]*mail.Address, error) {
	if e.header == nil {
		return nil, fmt.Errorf("No headers available")
	}
	if !AddressHeaders[strings.ToLower(key)] {
		return nil, fmt.Errorf("%s is not an address header", key)
	}

	str := decodeToUTF8Base64Header(e.header.Get(key))
	if str == "" {
		return nil, mail.ErrHeaderNotPresent
	}
	// These statements are handy for debugging ParseAddressList errors
	// fmt.Println("in:  ", m.header.Get(key))
	// fmt.Println("out: ", str)
	ret, err := mail.ParseAddressList(str)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// ReadEnvelope parses the content of the provided reader into an Envelope, downconverting HTML to
// plain text if needed, and sorting the attachments, inlines and other parts into their respective
// slices.
func ReadEnvelope(r io.Reader) (*Envelope, error) {
	// Read MIME parts from reader
	root, err := ReadParts(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to ReadParts: %v", err)
	}

	e := &Envelope{
		Root:   root,
		header: &root.Header,
	}

	if isMultipartMessage(root) {
		// Multi-part message (message with attachments, etc)
		if err := parseMultiPartBody(root, e); err != nil {
			return nil, err
		}
	} else {
		if isBinaryBody(root) {
			// Attachment only, no text
			if err := parseBinaryOnlyBody(root, e); err != nil {
				return nil, err
			}
		} else {
			// Only text, no attachments
			if err := parseTextOnlyBody(root, e); err != nil {
				return nil, err
			}
		}
	}

	// Copy part errors into Envelope
	if e.Root != nil {
		_ = e.Root.DepthMatchAll(func(part *Part) bool {
			// Using DepthMatchAll to traverse all parts, don't care about result
			for _, perr := range part.errors {
				e.Errors = append(e.Errors, &perr)
			}
			return false
		})
	}

	// Down-convert HTML to text if necessary
	if e.Text == "" && e.HTML != "" {
		e.IsTextFromHTML = true
		var err error
		if e.Text, err = html2text.FromString(e.HTML); err != nil {
			// Fail gently
			e.Text = ""
			return e, err
		}
	}

	return e, nil
}

// parseTextOnlyBody parses a plain text message in root that has MIME-like headers, but
// only contains a single part - no boundaries, etc.  The result is placed in e.
func parseTextOnlyBody(root *Part, e *Envelope) error {
	// Determine character set
	var charset string
	var isHTML bool
	if ctype := root.Header.Get("Content-Type"); ctype != "" {
		if mediatype, mparams, err := mime.ParseMediaType(ctype); err == nil {
			isHTML = (mediatype == "text/html")
			if mparams["charset"] != "" {
				charset = mparams["charset"]
			}
		}
	}

	// Setup character set transcoding
	var cr io.Reader
	if charset == "" {
		// No character set information, treat as UTF-8
		cr = root
	} else {
		// Convert text to UTF8 if content type specified a charset
		var err error
		if cr, err = newCharsetReader(charset, root); err != nil {
			return err
		}
	}

	// Read transcoded text
	bodyBytes, err := ioutil.ReadAll(cr)
	if err != nil {
		return err
	}
	if isHTML {
		rawHTML := string(bodyBytes)
		// Empty e.Text will trigger html2text conversion
		e.HTML = rawHTML
		if charset == "" {
			// Search for charset in HTML metadata
			if charset, _ = charsetFromHTMLString(rawHTML); charset != "" {
				// Found charset in HTML
				if convHTML, err := convertToUTF8String(charset, bodyBytes); err == nil {
					// Successful conversion
					e.HTML = convHTML
				} else {
					// Conversion failed
					root.addWarning(errorCharsetConversion, err.Error())
				}
			}
		}
	} else {
		e.Text = string(bodyBytes)
	}

	return nil
}

// parseBinaryOnlyBody parses a message where the only content is a binary attachment with no
// other parts. The result is placed in e.
func parseBinaryOnlyBody(root *Part, e *Envelope) error {
	// Determine mediatype
	ctype := root.Header.Get("Content-Type")
	mediatype, mparams, err := mime.ParseMediaType(ctype)
	if err != nil {
		mediatype = "attachment"
	}

	// TODO Find a way to share the duplicated code below with parseParts()
	// Determine and set headers for: content disposition, filename and character set
	disposition, dparams, err := mime.ParseMediaType(root.Header.Get("Content-Disposition"))
	if err == nil {
		// Disposition is optional
		root.SetDisposition(disposition)
		root.SetFileName(decodeHeader(dparams["filename"]))
	}
	if root.FileName() == "" && mparams["name"] != "" {
		root.SetFileName(decodeHeader(mparams["name"]))
	}
	if root.FileName() == "" && mparams["file"] != "" {
		root.SetFileName(decodeHeader(mparams["file"]))
	}
	if root.Charset() == "" {
		root.SetCharset(mparams["charset"])
	}

	// Add our part to the appropriate section of the Envelope
	e.Root = NewPart(nil, mediatype)

	if disposition == "inline" {
		e.Inlines = append(e.Inlines, root)
	} else {
		e.Attachments = append(e.Attachments, root)
	}

	return nil
}

// parseMultiPartBody parses a multipart message in root.  The result is placed in e.
func parseMultiPartBody(root *Part, e *Envelope) error {
	// Parse top-level multipart
	ctype := root.Header.Get("Content-Type")
	mediatype, params, err := mime.ParseMediaType(ctype)
	if err != nil {
		return fmt.Errorf("Unable to parse media type: %v", err)
	}
	if !strings.HasPrefix(mediatype, "multipart/") {
		return fmt.Errorf("Unknown mediatype: %v", mediatype)
	}
	boundary := params["boundary"]
	if boundary == "" {
		return fmt.Errorf("Unable to locate boundary param in Content-Type header")
	}

	// Locate text body
	if mediatype == "multipart/altern" {
		match := root.BreadthMatchFirst(func(p *Part) bool {
			return p.ContentType() == "text/plain" && p.Disposition() != "attachment"
		})
		if match != nil {
			var reader io.Reader
			if match.Charset() != "" {
				reader, err = newCharsetReader(match.Charset(), match)
				if err != nil {
					return err
				}
			} else {
				reader = match
			}
			allBytes, ioerr := ioutil.ReadAll(reader)
			if ioerr != nil {
				return ioerr
			}
			e.Text += string(allBytes)
		}
	} else {
		// multipart is of a mixed type
		match := root.DepthMatchAll(func(p *Part) bool {
			return p.ContentType() == "text/plain" && p.Disposition() != "attachment"
		})
		for i, m := range match {
			if i > 0 {
				e.Text += "\n--\n"
			}
			var reader io.Reader
			if m.Charset() != "" {
				reader, err = newCharsetReader(m.Charset(), m)
				if err != nil {
					return err
				}
			} else {
				reader = m
			}
			allBytes, ioerr := ioutil.ReadAll(reader)
			if ioerr != nil {
				return ioerr
			}
			e.Text += string(allBytes)
		}
	}

	// Locate HTML body
	match := root.BreadthMatchFirst(func(p *Part) bool {
		return p.ContentType() == "text/html" && p.Disposition() != "attachment"
	})
	if match != nil {
		var reader io.Reader
		if match.Charset() != "" {
			reader, err = newCharsetReader(match.Charset(), match)
			if err != nil {
				return err
			}
		} else {
			reader = match
		}
		allBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		e.HTML += string(allBytes)
	}

	// Locate attachments
	e.Attachments = root.BreadthMatchAll(func(p *Part) bool {
		return p.Disposition() == "attachment" || p.ContentType() == "application/octet-stream"
	})

	// Locate inlines
	e.Inlines = root.BreadthMatchAll(func(p *Part) bool {
		return p.Disposition() == "inline"
	})

	// Locate others parts not considered in attachments or inlines
	e.OtherParts = root.BreadthMatchAll(func(p *Part) bool {
		if strings.HasPrefix(p.ContentType(), "multipart/") {
			return false
		}
		if p.Disposition() != "" {
			return false
		}
		if p.ContentType() == "application/octet-stream" {
			return false
		}
		return p.ContentType() != "text/plain" && p.ContentType() != "text/html"
	})

	return nil
}

// isMultipartMessage returns true if the message has a recognized multipart Content-Type header.
func isMultipartMessage(root *Part) bool {
	// Parse top-level multipart
	ctype := root.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ctype)
	if err != nil {
		return false
	}
	// According to rfc2046#section-5.1.7 all other multipart should
	// be treated as multipart/mixed
	return strings.HasPrefix(mediatype, "multipart/")
}

// isAttachment returns true, if the given header defines an attachment.  First it checks if the
// Content-Disposition header defines an attachement or inline attachment. If this test is false,
// the Content-Type header is checked for attachment, but not inline.  Email clients use inline for
// their text bodies.
//
// Valid Attachment-Headers:
//
//  - Content-Disposition: attachment; filename="frog.jpg"
//  - Content-Disposition: inline; filename="frog.jpg"
//  - Content-Type: attachment; filename="frog.jpg"
func isAttachment(header textproto.MIMEHeader) bool {
	mediatype, _, _ := mime.ParseMediaType(header.Get("Content-Disposition"))
	if strings.ToLower(mediatype) == "attachment" ||
		strings.ToLower(mediatype) == "inline" {
		return true
	}

	mediatype, _, _ = mime.ParseMediaType(header.Get("Content-Type"))
	if strings.ToLower(mediatype) == "attachment" {
		return true
	}

	return false
}

// isPlain returns true, if the the MIME headers define a valid 'text/plain' or 'text/html' part. If
// the emptyContentTypeIsPlain argument is set to true, a missing Content-Type header will result in
// a positive plain part detection.
func isPlain(header textproto.MIMEHeader, emptyContentTypeIsPlain bool) bool {
	ctype := header.Get("Content-Type")
	if ctype == "" && emptyContentTypeIsPlain {
		return true
	}

	mediatype, _, err := mime.ParseMediaType(ctype)
	if err != nil {
		return false
	}
	switch mediatype {
	case "text/plain",
		"text/html":
		return true
	}

	return false
}

// isBinaryBody returns true if the mail header defines a binary body.
func isBinaryBody(root *Part) bool {
	if isAttachment(root.Header) {
		return true
	}

	return !isPlain(root.Header, true)
}
