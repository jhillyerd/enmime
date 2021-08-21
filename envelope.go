package enmime

import (
	"fmt"
	"io"
	"mime"
	"net/mail"
	"net/textproto"
	"strings"

	"github.com/jaytaylor/html2text"
	"github.com/jhillyerd/enmime/internal/coding"
	"github.com/jhillyerd/enmime/mediatype"
	"github.com/pkg/errors"
)

// Envelope is a simplified wrapper for MIME email messages.
type Envelope struct {
	Text        string  // The plain text portion of the message
	HTML        string  // The HTML portion of the message
	Root        *Part   // The top-level Part
	Attachments []*Part // All parts having a Content-Disposition of attachment
	Inlines     []*Part // All parts having a Content-Disposition of inline
	// All non-text parts that were not placed in Attachments or Inlines, such as multipart/related
	// content.
	OtherParts []*Part
	Errors     []*Error              // Errors encountered while parsing
	header     *textproto.MIMEHeader // Header from original message
}

// GetHeaderKeys returns a list of header keys seen in this message. Get
// individual headers with `GetHeader(name)`
func (e *Envelope) GetHeaderKeys() (headers []string) {
	if e.header == nil {
		return
	}
	for key := range *e.header {
		headers = append(headers, key)
	}
	return headers
}

// GetHeader processes the specified header for RFC 2047 encoded words and returns the result as a
// UTF-8 string
func (e *Envelope) GetHeader(name string) string {
	if e.header == nil {
		return ""
	}
	return coding.DecodeExtHeader(e.header.Get(name))
}

// GetHeaderValues processes the specified header for RFC 2047 encoded words and returns all existing
// values as a list of UTF-8 strings
func (e *Envelope) GetHeaderValues(name string) []string {
	if e.header == nil {
		return []string{}
	}

	rawValues := (*e.header)[textproto.CanonicalMIMEHeaderKey(name)]
	var values []string
	for _, v := range rawValues {
		values = append(values, coding.DecodeExtHeader(v))
	}
	return values
}

// SetHeader sets given header name to the given value.
// If the header exists already, all existing values are replaced.
func (e *Envelope) SetHeader(name string, value []string) error {
	if name == "" {
		return fmt.Errorf("provide non-empty header name")
	}

	for i, v := range value {
		if i == 0 {
			e.header.Set(name, mime.BEncoding.Encode("utf-8", v))
			continue
		}
		e.header.Add(name, mime.BEncoding.Encode("utf-8", v))
	}
	return nil
}

// AddHeader appends given header value to header name without changing existing values.
// If the header does not exist already, it will be created.
func (e *Envelope) AddHeader(name string, value string) error {
	if name == "" {
		return fmt.Errorf("provide non-empty header name")
	}

	e.header.Add(name, mime.BEncoding.Encode("utf-8", value))
	return nil
}

// DeleteHeader deletes given header.
func (e *Envelope) DeleteHeader(name string) error {
	if name == "" {
		return fmt.Errorf("provide non-empty header name")
	}

	e.header.Del(name)
	return nil
}

// AddressList returns a mail.Address slice with RFC 2047 encoded names converted to UTF-8
func (e *Envelope) AddressList(key string) ([]*mail.Address, error) {
	if e.header == nil {
		return nil, fmt.Errorf("no headers available")
	}
	if !AddressHeaders[strings.ToLower(key)] {
		return nil, fmt.Errorf("%s is not an address header", key)
	}

	str := decodeToUTF8Base64Header(e.header.Get(key))

	// These statements are handy for debugging ParseAddressList errors
	// fmt.Println("in:  ", m.header.Get(key))
	// fmt.Println("out: ", str)
	ret, err := mail.ParseAddressList(str)
	if err != nil {
		switch err.Error() {
		case "mail: expected comma":
			return mail.ParseAddressList(ensureCommaDelimitedAddresses(str))
		case "mail: no address":
			return nil, mail.ErrHeaderNotPresent
		}
		return nil, err
	}
	return ret, nil
}

// Clone returns a clone of the current Envelope
func (e *Envelope) Clone() *Envelope {
	if e == nil {
		return nil
	}

	newEnvelope := &Envelope{
		e.Text,
		e.HTML,
		e.Root.Clone(nil),
		e.Attachments,
		e.Inlines,
		e.OtherParts,
		e.Errors,
		e.header,
	}
	return newEnvelope
}

// ReadEnvelope is a wrapper around ReadParts and EnvelopeFromPart.  It parses the content of the
// provided reader into an Envelope, downconverting HTML to plain text if needed, and sorting the
// attachments, inlines and other parts into their respective slices. Errors are collected from all
// Parts and placed into the Envelope.Errors slice.
func ReadEnvelope(r io.Reader) (*Envelope, error) {
	// Read MIME parts from reader
	root, err := ReadParts(r)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to ReadParts")
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
			if root.Disposition == cdInline {
				e.Inlines = append(e.Inlines, root)
			} else {
				e.Attachments = append(e.Attachments, root)
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
			e.Text = "" // Down-conversion shouldn't fail
			p := e.Root.BreadthMatchFirst(matchHTMLBodyPart)
			p.addError(ErrorPlainTextFromHTML, "Failed to downconvert HTML: %v", err)
		}
	}

	// Copy part errors into Envelope.
	if e.Root != nil {
		_ = e.Root.DepthMatchAll(func(part *Part) bool {
			// Using DepthMatchAll to traverse all parts, don't care about result.
			for i := range part.Errors {
				// Range index is needed to get the correct address, because range value points to
				// a locally scoped variable.
				e.Errors = append(e.Errors, part.Errors[i])
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
		if mediatype, mparams, _, err := mediatype.Parse(ctype); err == nil {
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
			if charset = coding.FindCharsetInHTML(rawHTML); charset != "" {
				// Found charset in HTML
				if convHTML, err := coding.ConvertToUTF8String(charset, root.Content); err == nil {
					// Successful conversion
					e.HTML = convHTML
				} else {
					// Conversion failed
					root.addWarning(ErrorCharsetConversion, err.Error())
				}
			}
			// Converted from charset in HTML
			return nil
		}
	} else {
		e.Text = string(root.Content)
	}

	return nil
}

// parseMultiPartBody parses a multipart message in root.  The result is placed in e.
func parseMultiPartBody(root *Part, e *Envelope) error {
	// Parse top-level multipart
	ctype := root.Header.Get(hnContentType)
	mediatype, params, _, err := mediatype.Parse(ctype)
	if err != nil {
		return fmt.Errorf("unable to parse media type: %v", err)
	}
	if !strings.HasPrefix(mediatype, ctMultipartPrefix) {
		return fmt.Errorf("unknown mediatype: %v", mediatype)
	}
	boundary := params[hpBoundary]
	if boundary == "" {
		return fmt.Errorf("unable to locate boundary param in Content-Type header")
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
	p := root.DepthMatchFirst(matchHTMLBodyPart)
	if p != nil {
		e.HTML += string(p.Content)
	}

	// Locate attachments
	e.Attachments = root.BreadthMatchAll(func(p *Part) bool {
		return p.Disposition == cdAttachment || p.ContentType == ctAppOctetStream
	})

	// Locate inlines
	e.Inlines = root.BreadthMatchAll(func(p *Part) bool {
		return p.Disposition == cdInline && !strings.HasPrefix(p.ContentType, ctMultipartPrefix)
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

// Used by AddressList to ensure that address lists are properly delimited
func ensureCommaDelimitedAddresses(s string) string {
	// This normalizes the whitespace, but may interfere with CFWS (comments with folding whitespace)
	// RFC-5322 3.4.0:
	//      because some legacy implementations interpret the comment,
	//      comments generally SHOULD NOT be used in address fields
	//      to avoid confusing such implementations.
	s = strings.Join(strings.Fields(s), " ")

	inQuotes := false
	inDomain := false
	escapeSequence := false
	sb := strings.Builder{}
	for _, r := range s {
		if escapeSequence {
			escapeSequence = false
			sb.WriteRune(r)
			continue
		}
		if r == '"' {
			inQuotes = !inQuotes
			sb.WriteRune(r)
			continue
		}
		if inQuotes {
			if r == '\\' {
				escapeSequence = true
				sb.WriteRune(r)
				continue
			}
		} else {
			if r == '@' {
				inDomain = true
				sb.WriteRune(r)
				continue
			}
			if inDomain {
				if r == ';' {
					sb.WriteRune(r)
					break
				}
				if r == ',' {
					inDomain = false
					sb.WriteRune(r)
					continue
				}
				if r == ' ' {
					inDomain = false
					sb.WriteRune(',')
					sb.WriteRune(r)
					continue
				}
			}
		}
		sb.WriteRune(r)
	}
	return sb.String()
}
