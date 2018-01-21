package enmime

import (
	"bytes"
	"errors"
	"net/smtp"
	"reflect"
	"strings"
	"time"
)

// MailBuilder facilitates the easy construction of a MIME message.  Each manipulation method
// returns a copy of the receiver struct.  It can be considered immutable if the caller does not
// modify the string and byte slices passed in.  Immutability allows the headers or entire message
// to be reused across multiple threads.
type MailBuilder struct {
	to, cc, bcc   []string
	from, subject string
	date          time.Time
	text, html    []byte
}

// Builder returns an empty MailBuilder struct
func Builder() *MailBuilder {
	return &MailBuilder{}
}

// From returns a copy of MailBuilder with the specified From header
func (p *MailBuilder) From(from string) *MailBuilder {
	c := *p
	c.from = from
	return &c
}

// Subject returns a copy of MailBuilder with the specified Subject header
func (p *MailBuilder) Subject(subject string) *MailBuilder {
	c := *p
	c.subject = subject
	return &c
}

// To returns a copy of MailBuilder with the specified To header
func (p *MailBuilder) To(to []string) *MailBuilder {
	c := *p
	c.to = to
	return &c
}

// CC returns a copy of MailBuilder with the specified Cc header
func (p *MailBuilder) CC(cc []string) *MailBuilder {
	c := *p
	c.cc = cc
	return &c
}

// BCC returns a copy of MailBuilder with the specified recipients added to the blind CC list.  This
// method only has an effect if the Send method is used to transmit the message, there is no effect
// on the parts returned by the Build()
func (p *MailBuilder) BCC(bcc []string) *MailBuilder {
	c := *p
	c.bcc = bcc
	return &c
}

// Text returns a copy of MailBuilder that will use the provided bytes for its text/plain Part
func (p *MailBuilder) Text(body []byte) *MailBuilder {
	c := *p
	c.text = body
	return &c
}

// HTML returns a copy of MailBuilder that will use the provided bytes for its text/html Part
func (p *MailBuilder) HTML(body []byte) *MailBuilder {
	c := *p
	c.html = body
	return &c
}

// Build performs some basic validations, then constructs a tree of Part structs from the configured
// MailBuilder.  It will set the Date header to now if it was not explicitly set.
func (p *MailBuilder) Build() (*Part, error) {
	// Validations
	if p.from == "" {
		return nil, errors.New("from not set")
	}
	if p.subject == "" {
		return nil, errors.New("subject not set")
	}
	if len(p.to)+len(p.cc)+len(p.bcc) == 0 {
		return nil, errors.New("no recipients (to, cc, bcc) set")
	}
	var root *Part
	if p.text != nil && p.html != nil {
		// Multipart
		root = NewPart(nil, ctMultipartAltern)
		t := NewPart(root, ctTextPlain)
		root.FirstChild = t
		t.Content = p.text
		t.Charset = "utf-8"
		h := NewPart(root, ctTextHTML)
		t.NextSibling = h
		h.Content = p.html
		h.Charset = "utf-8"
	} else if p.html != nil {
		// HTML only
		root = NewPart(nil, ctTextHTML)
		root.Content = p.html
		root.Charset = "utf-8"
	} else {
		// Default to text only, even if empty
		root = NewPart(nil, ctTextPlain)
		root.Content = p.text
		root.Charset = "utf-8"
	}
	// Headers
	h := root.Header
	h.Set("From", p.from)
	h.Set("Subject", p.subject)
	if len(p.to) > 0 {
		h.Set("To", strings.Join(p.to, ", "))
	}
	if len(p.cc) > 0 {
		h.Set("Cc", strings.Join(p.cc, ", "))
	}
	date := p.date
	if date.IsZero() {
		date = time.Now()
	}
	h.Set("Date", date.Format(time.RFC1123Z))
	return root, nil
}

// Send encodes the message and sends it via the SMTP server specified by addr.  Send uses
// net/smtp.SendMail, and accepts the same authentication parameters.
func (p *MailBuilder) Send(addr string, a smtp.Auth) error {
	buf := &bytes.Buffer{}
	root, err := p.Build()
	if err != nil {
		return err
	}
	err = root.Encode(buf)
	if err != nil {
		return err
	}
	recips := make([]string, 0, len(p.to)+len(p.cc)+len(p.bcc))
	recips = append(recips, p.to...)
	recips = append(recips, p.cc...)
	recips = append(recips, p.bcc...)
	return smtp.SendMail(addr, a, p.from, recips, buf.Bytes())
}

// Equals uses the reflect package to test two MailBuilder structs for equality, primarily for unit
// tests
func (p *MailBuilder) Equals(o *MailBuilder) bool {
	if p == nil {
		return o == nil
	}
	return reflect.DeepEqual(p, o)
}
