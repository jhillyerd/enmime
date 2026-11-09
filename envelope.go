package enmime

import (
	"io"
	"mime"
	"net/mail"
	"net/textproto"
	"sort"
	"strings"
	"time"

	"github.com/inbucket/html2text"
	"github.com/jhillyerd/enmime/v2/internal/coding"
	inttp "github.com/jhillyerd/enmime/v2/internal/textproto"
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

	rawValues := (*e.header)[inttp.CanonicalEmailMIMEHeaderKey(name)]
	values := make([]string, 0, len(rawValues))
	for _, v := range rawValues {
		values = append(values, coding.DecodeExtHeader(v))
	}
	return values
}

// SetHeader sets given header name to the given value.
// If the header exists already, all existing values are replaced.
func (e *Envelope) SetHeader(name string, value []string) error {
	if name == "" {
		return errors.New("provide non-empty header name")
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
		return errors.New("provide non-empty header name")
	}

	e.header.Add(name, mime.BEncoding.Encode("utf-8", value))
	return nil
}

// DeleteHeader deletes given header.
func (e *Envelope) DeleteHeader(name string) error {
	if name == "" {
		return errors.New("provide non-empty header name")
	}

	e.header.Del(name)
	return nil
}

// SetAddressHeader sets the named address header to addrs, formatted as an RFC 5322 address list.
// If the header exists already, all existing values are replaced; passing a nil or empty addrs
// slice deletes the header instead of writing an empty value.
//
// Names containing non-ASCII characters are emitted as RFC 2047 encoded-words, so the header stays
// ASCII-safe on the wire while remaining a parseable address list.  Pure-ASCII names are written
// unchanged, keeping the familiar `Bob <bob@example.com>` form.  Unlike SetHeader, which encodes an
// entire value as one encoded-word, only the display name of each address is encoded; the address
// list itself is never wrapped inside an encoded-word, so AddressList and ParseAddressList can read
// the result back.
//
// Pass names as plain UTF-8.  A name that is already an RFC 2047 encoded-word is passed through
// unchanged, so pre-encoded input from another source is not double-encoded.
func (e *Envelope) SetAddressHeader(name string, addrs []*mail.Address) error {
	if name == "" {
		return errors.New("provide non-empty header name")
	}
	if !AddressHeaders[strings.ToLower(name)] {
		return errors.Errorf("%s is not an address header, expected one of: %s", name,
			strings.Join(sortedAddressHeaders(), ", "))
	}
	if e.header == nil {
		return errors.New("no headers available")
	}

	var buf strings.Builder
	count := 0
	for _, addr := range addrs {
		if addr == nil {
			continue
		}
		// A value containing CR or LF would be written verbatim by Part.encodeHeader when the parser
		// runs in raw-content mode, letting it inject headers into the transmitted message.
		if err := validateSendAddress(addr.Address); err != nil {
			return err
		}
		if count > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(formatAddress(addr))
		count++
	}
	if count == 0 {
		e.header.Del(name)
		return nil
	}

	e.header.Set(name, buf.String())
	return nil
}

// formatAddress renders a single address for an address header.  A Name holding non-ASCII bytes is
// emitted as an RFC 2047 encoded-word ahead of the addr-spec, which is what mail.Address.String
// does; a Name that is plain ASCII is written verbatim when it is a legal RFC 5322 phrase, so the
// common case reads Bob <bob@example.com>.  Names that are neither are quoted by String.
func formatAddress(addr *mail.Address) string {
	if addr.Name != "" && !addressNameNeedsEncoding(addr.Name) && addressNameIsPhrase(addr.Name) {
		return addr.Name + " " + (&mail.Address{Address: addr.Address}).String()
	}
	return addr.String()
}

// addressNameNeedsEncoding reports whether name cannot be written on the wire as-is, either because
// it holds non-ASCII bytes or because it holds control characters.
func addressNameNeedsEncoding(name string) bool {
	for _, r := range name {
		if (r < ' ' || r > '~') && r != '\t' {
			return true
		}
	}
	return false
}

// addressNameIsPhrase reports whether name may stand alone as an RFC 5322 phrase, without being
// quoted or encoded.  RFC 5322 3.2.4 defines phrase = (atext / SP)*, so only printable ASCII is
// allowed, and none of the specials ()<>[]:;@\,." may appear.
func addressNameIsPhrase(name string) bool {
	if name == "" || name[0] == ' ' || name[len(name)-1] == ' ' {
		return false
	}
	for _, r := range name {
		if r == ' ' {
			continue
		}
		if r < '!' || r > '~' || strings.ContainsRune("()<>[]:;@\\,.\"", r) {
			return false
		}
	}
	return true
}

// sortedAddressHeaders returns the keys of AddressHeaders in sorted order, for error messages.
func sortedAddressHeaders() []string {
	keys := make([]string, 0, len(AddressHeaders))
	for k := range AddressHeaders {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// AddressList returns a mail.Address slice with RFC 2047 encoded names converted to UTF-8
func (e *Envelope) AddressList(key string) ([]*mail.Address, error) {
	if e.header == nil {
		return nil, errors.New("no headers available")
	}
	if !AddressHeaders[strings.ToLower(key)] {
		return nil, errors.Errorf("%s is not an address header", key)
	}

	return ParseAddressList(e.header.Get(key))
}

// Date parses the Date header field.
func (e *Envelope) Date() (time.Time, error) {
	hdr := e.GetHeader("Date")
	if hdr == "" {
		return time.Time{}, mail.ErrHeaderNotPresent
	}
	return mail.ParseDate(hdr)
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
// Uses default parser.
func ReadEnvelope(r io.Reader) (*Envelope, error) {
	return defaultParser.ReadEnvelope(r)
}

// ReadEnvelope is the same as ReadEnvelope, but respects parser configurations.
func (p Parser) ReadEnvelope(r io.Reader) (*Envelope, error) {
	// Read MIME parts from reader
	root, err := p.ReadParts(r)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to ReadParts")
	}

	e, err := p.EnvelopeFromPart(root)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to EnvelopeFromPart")
	}

	err = e.GatherNestedErrors()
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to GatherNestedErrors")
	}

	return e, nil
}

// EnvelopeFromPart uses the provided Part tree to build an Envelope, downconverting HTML to plain
// text if needed, and sorting the attachments, inlines and other parts into their respective
// slices.  Gather errors from all parts with Envelope.GatherNestedErrors().
func EnvelopeFromPart(root *Part) (*Envelope, error) {
	return defaultParser.EnvelopeFromPart(root)
}

// EnvelopeFromPart is the same as EnvelopeFromPart, but respects parser configurations.
func (p Parser) EnvelopeFromPart(root *Part) (*Envelope, error) {
	e := &Envelope{
		Root:   root,
		header: &root.Header,
	}

	if detectMultipartMessage(root, p.multipartWOBoundaryAsSinglePart) {
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
			parseTextOnlyBody(root, e)
		}
	}

	// Down-convert HTML to text if necessary
	if e.Text == "" && e.HTML != "" {
		// We always warn when this happens
		e.Root.addWarning(ErrorPlainTextFromHTML, "Message did not contain a text/plain part")

		if !p.disableTextConversion {
			var err error
			if e.Text, err = html2text.FromString(e.HTML); err != nil {
				e.Text = "" // Down-conversion shouldn't fail
				p := e.Root.BreadthMatchFirst(matchHTMLBodyPart)
				p.addErrorf(ErrorPlainTextFromHTML, "Failed to downconvert HTML: %v", err)
			}
		}
	}

	return e, nil
}

// GatherNestedErrors gathers errors from the entire part tree (including the root) and adds them to the Envelope.
func (e *Envelope) GatherNestedErrors() error {
	// Copy part errors from all nested/child/sibling parts into Envelope.
	if e.Root != nil {
		_ = e.Root.DepthMatchAll(func(part *Part) bool {
			// Using DepthMatchAll to traverse all parts, don't care about result.
			e.Errors = append(e.Errors, part.Errors...)
			return false
		})
	}

	return nil
}

// parseTextOnlyBody parses a plain text message in root that has MIME-like headers, but
// only contains a single part - no boundaries, etc.  The result is placed in e.
func parseTextOnlyBody(root *Part, e *Envelope) {
	// Determine character set
	var charset string
	var isHTML bool
	if ctype := root.Header.Get(hnContentType); ctype != "" {
		if mediatype, mparams, _, err := root.parseMediaType(ctype); err == nil {
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
		}
	} else {
		e.Text = string(root.Content)
	}
}

// parseMultiPartBody parses a multipart message in root.  The result is placed in e.
func parseMultiPartBody(root *Part, e *Envelope) error {
	// Parse top-level multipart
	ctype := root.Header.Get(hnContentType)
	mediatype, params, _, err := root.parseMediaType(ctype)
	if err != nil {
		return errors.WithMessage(err, "unable to parse media type")
	}
	if !strings.HasPrefix(mediatype, ctMultipartPrefix) {
		return errors.Errorf("unknown mediatype: %v", mediatype)
	}
	boundary := params[hpBoundary]
	if boundary == "" {
		return errors.New("unable to locate boundary param in Content-Type header")
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
