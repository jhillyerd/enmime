package enmime

import (
	"fmt"
	"io"
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

	if detectMultipartMessage(root) {
		// Multi-part message (message with attachments, etc)
		if err := parseMultiPartBody(root, e); err != nil {
			return nil, err
		}
	} else {
		if detectBinaryBody(root) {
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
			ErrorPlainTextFromHTML,
			"Message did not contain a text/plain part")
		var err error
		if e.Text, err = html2text.FromString(e.HTML); err != nil {
			// Downcoversion shouldn't fail
			e.Text = ""
			p := e.Root.BreadthMatchFirst(matchHTMLBodyPart)
			p.addError(
				ErrorPlainTextFromHTML,
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
	if isHTML {
		rawHTML := string(root.Content)
		// Note: Empty e.Text will trigger html2text conversion
		e.HTML = rawHTML
		if charset == "" {
			// Search for charset in HTML metadata
			if charset = findCharsetInHTML(rawHTML); charset != "" {
				// Found charset in HTML
				if convHTML, err := convertToUTF8String(charset, root.Content); err == nil {
					// Successful conversion
					e.HTML = convHTML
				} else {
					// Conversion failed
					root.addWarning(ErrorCharsetConversion, err.Error())
				}
			}
		}
	} else {
		e.Text = string(root.Content)
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

	// Add header from binary only part
	e.Root.Header = root.Header

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
			e.Text = string(p.Content)
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
			e.Text += string(p.Content)
		}
	}

	// Locate HTML body
	p := root.BreadthMatchFirst(matchHTMLBodyPart)
	if p != nil {
		e.HTML += string(p.Content)
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

// Used by Part matchers to locate the HTML body.  Not inlined because it's used in multiple places.
func matchHTMLBodyPart(p *Part) bool {
	return p.ContentType == ctTextHTML && p.Disposition != cdAttachment
}
