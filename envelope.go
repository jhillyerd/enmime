package enmime

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/mail"
	"net/textproto"
	"strings"

	"github.com/jaytaylor/html2text"
)

// Envelope is a simplified wrapper for MIME email messages.
type Envelope struct {
	Text        string                // The plain text portion of the message
	HTML        string                // The HTML portion of the message
	Root        *Part                 // The top-level Part
	Attachments []*Part               // All parts having a Content-Disposition of attachment
	Inlines     []*Part               // All parts having a Content-Disposition of inline
	OtherParts  []*Part               // All parts not in Attachments and Inlines
	Errors      []*Error              // Errors encountered while parsing
	header      *textproto.MIMEHeader // Header from original message
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

// ReadEnvelope is a wrapper around ReadParts and EnvelopeFromPart.  It parses the content of the
// provided reader into an Envelope, downconverting HTML to plain text if needed, and sorting the
// attachments, inlines and other parts into their respective slices. Errors are collected from all
// Parts and placed into the Envelope.Errors slice.
func ReadEnvelope(r io.Reader) (*Envelope, error) {
	// Read MIME parts from reader
	root, err := ReadParts(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to ReadParts: %v", err)
	}
	return EnvelopeFromPart(root)
}

// EnvelopeFromPart uses the provided Part tree to build an Envelope, downconverting HTML to plain
// text if needed, and sorting the attachments, inlines and other parts into their respective
// slices.  Errors are collected from all Parts and placed into the Envelopes Errors slice.
func EnvelopeFromPart(root *Part) (*Envelope, error) {
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

	// Down-convert HTML to text if necessary
	if e.Text == "" && e.HTML != "" {
		// We always warn when this happens
		e.Root.addWarning(
			errorPlainTextFromHTML,
			"Message did not contain a text/plain part")
		var err error
		if e.Text, err = html2text.FromString(e.HTML); err != nil {
			// Downcoversion shouldn't fail
			e.Text = ""
			p := e.Root.BreadthMatchFirst(matchHTMLBodyPart)
			p.addError(
				errorPlainTextFromHTML,
				"Failed to downconvert HTML: %v",
				err)
		}
	}

	// Copy part errors into Envelope
	if e.Root != nil {
		_ = e.Root.DepthMatchAll(func(part *Part) bool {
			// Using DepthMatchAll to traverse all parts, don't care about result
			for i := range part.Errors {
				// Index is required here to get the correct address, &value from range
				// points to a locally scoped variable
				e.Errors = append(e.Errors, &part.Errors[i])
			}
			return false
		})
	}

	return e, nil
}

// parseTextOnlyBody parses a plain text message in root that has MIME-like headers, but
// only contains a single part - no boundaries, etc.  The result is placed in e.
func parseTextOnlyBody(root *Part, e *Envelope) error {
	// Determine character set
	var charset string
	var isHTML bool
	if ctype := root.Header.Get(hnContentType); ctype != "" {
		if mediatype, mparams, err := parseMediaType(ctype); err == nil {
			isHTML = (mediatype == ctTextHTML)
			if mparams[hpCharset] != "" {
				charset = mparams[hpCharset]
			}
		}
	}

	// Read transcoded text
	bodyBytes, err := ioutil.ReadAll(root)
	if err != nil {
		return err
	}
	if isHTML {
		rawHTML := string(bodyBytes)
		// Note: Empty e.Text will trigger html2text conversion
		e.HTML = rawHTML
		if charset == "" {
			// Search for charset in HTML metadata
			if charset = findCharsetInHTML(rawHTML); charset != "" {
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
	ctype := root.Header.Get(hnContentType)
	mediatype, mparams, err := parseMediaType(ctype)
	if err != nil {
		mediatype = cdAttachment
	}

	// Determine and set headers for: content disposition, filename and character set
	root.setupContentHeaders(mparams)

	// Add our part to the appropriate section of the Envelope
	e.Root = NewPart(nil, mediatype)

	if root.Disposition == cdInline {
		e.Inlines = append(e.Inlines, root)
	} else {
		e.Attachments = append(e.Attachments, root)
	}

	return nil
}

// parseMultiPartBody parses a multipart message in root.  The result is placed in e.
func parseMultiPartBody(root *Part, e *Envelope) error {
	// Parse top-level multipart
	ctype := root.Header.Get(hnContentType)
	mediatype, params, err := parseMediaType(ctype)
	if err != nil {
		return fmt.Errorf("Unable to parse media type: %v", err)
	}
	if !strings.HasPrefix(mediatype, ctMultipartPrefix) {
		return fmt.Errorf("Unknown mediatype: %v", mediatype)
	}
	boundary := params[hpBoundary]
	if boundary == "" {
		return fmt.Errorf("Unable to locate boundary param in Content-Type header")
	}

	// Locate text body
	if mediatype == ctMultipartAltern {
		p := root.BreadthMatchFirst(func(p *Part) bool {
			return p.ContentType == ctTextPlain && p.Disposition != cdAttachment
		})
		if p != nil {
			allBytes, ioerr := ioutil.ReadAll(p)
			if ioerr != nil {
				return ioerr
			}
			e.Text = string(allBytes)
		}
	} else {
		// multipart is of a mixed type
		parts := root.DepthMatchAll(func(p *Part) bool {
			return p.ContentType == ctTextPlain && p.Disposition != cdAttachment
		})
		for i, p := range parts {
			if i > 0 {
				e.Text += "\n--\n"
			}
			allBytes, ioerr := ioutil.ReadAll(p)
			if ioerr != nil {
				return ioerr
			}
			e.Text += string(allBytes)
		}
	}

	// Locate HTML body
	p := root.BreadthMatchFirst(matchHTMLBodyPart)
	if p != nil {
		allBytes, ioerr := ioutil.ReadAll(p)
		if ioerr != nil {
			return ioerr
		}
		e.HTML += string(allBytes)
	}

	// Locate attachments
	e.Attachments = root.BreadthMatchAll(func(p *Part) bool {
		return p.Disposition == cdAttachment || p.ContentType == ctAppOctetStream
	})

	// Locate inlines
	e.Inlines = root.BreadthMatchAll(func(p *Part) bool {
		return p.Disposition == cdInline
	})

	// Locate others parts not considered in attachments or inlines
	e.OtherParts = root.BreadthMatchAll(func(p *Part) bool {
		if strings.HasPrefix(p.ContentType, ctMultipartPrefix) {
			return false
		}
		if p.Disposition != "" {
			return false
		}
		if p.ContentType == ctAppOctetStream {
			return false
		}
		return p.ContentType != ctTextPlain && p.ContentType != ctTextHTML
	})

	return nil
}

// isMultipartMessage returns true if the message has a recognized multipart Content-Type header.
func isMultipartMessage(root *Part) bool {
	// Parse top-level multipart
	ctype := root.Header.Get(hnContentType)
	mediatype, _, err := parseMediaType(ctype)
	if err != nil {
		return false
	}
	// According to rfc2046#section-5.1.7 all other multipart should
	// be treated as multipart/mixed
	return strings.HasPrefix(mediatype, ctMultipartPrefix)
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
	mediatype, _, _ := parseMediaType(header.Get(hnContentDisposition))
	if strings.ToLower(mediatype) == cdAttachment ||
		strings.ToLower(mediatype) == cdInline {
		return true
	}

	mediatype, _, _ = parseMediaType(header.Get(hnContentType))
	if strings.ToLower(mediatype) == cdAttachment {
		return true
	}

	return false
}

// isPlain returns true, if the the MIME headers define a valid 'text/plain' or 'text/html' part. If
// the emptyContentTypeIsPlain argument is set to true, a missing Content-Type header will result in
// a positive plain part detection.
func isPlain(header textproto.MIMEHeader, emptyContentTypeIsPlain bool) bool {
	ctype := header.Get(hnContentType)
	if ctype == "" && emptyContentTypeIsPlain {
		return true
	}

	mediatype, _, err := parseMediaType(ctype)
	if err != nil {
		return false
	}
	switch mediatype {
	case ctTextPlain, ctTextHTML:
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

// Used by Part matchers to locate the HTML body.  Not inlined because it's used in multiple places.
func matchHTMLBodyPart(p *Part) bool {
	return p.ContentType == ctTextHTML && p.Disposition != cdAttachment
}
