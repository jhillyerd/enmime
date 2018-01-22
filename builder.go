package enmime

import (
	"bytes"
	"errors"
	"net/smtp"
	"net/textproto"
	"reflect"
	"strings"
	"time"
)

// MailBuilder facilitates the easy construction of a MIME message.  Each manipulation method
// returns a copy of the receiver struct.  It can be considered immutable if the caller does not
// modify the string and byte slices passed in.  Immutability allows the headers or entire message
// to be reused across multiple threads.
type MailBuilder struct {
	to, cc, bcc          []string
	from, subject        string
	date                 time.Time
	text, html           []byte
	inlines, attachments []*Part
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

// AddAttachment returns a copy of MailBuilder that includes the specified attachment
func (p *MailBuilder) AddAttachment(b []byte, contentType string, fileName string) *MailBuilder {
	part := NewPart(nil, contentType)
	part.Content = b
	part.FileName = fileName
	c := *p
	c.attachments = append(c.attachments, part)
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
	/**
	 * Fully loaded structure; the presence of text, html, inlines, and attachments will determine
	 * how much is necessary:
	 *
	 * multipart/mixed
	 * |- multipart/related
	 * |  |- multipart/alternative
	 * |  |  |- text/plain
	 * |  |  `- text/html
	 * |  `- inlines..
	 * `- attachments..
	 *
	 * We build this tree starting at the leaves, re-rooting as needed.
	 */
	var root, part *Part
	if p.text != nil || p.html == nil {
		root = NewPart(nil, ctTextPlain)
		root.Content = p.text
		root.Charset = "utf-8"
	}
	if p.html != nil {
		part = NewPart(nil, ctTextHTML)
		part.Content = p.html
		part.Charset = "utf-8"
		if root == nil {
			root = part
		} else {
			root.NextSibling = part
		}
	}
	if p.text != nil && p.html != nil {
		// Wrap Text & HTML bodies
		part = root
		root = NewPart(nil, ctMultipartAltern)
		root.AddChild(part)
	}
	if len(p.attachments) > 0 {
		part = root
		root = NewPart(nil, ctMultipartMixed)
		root.AddChild(part)
		for _, ap := range p.attachments {
			// Copy attachment Part to isolate mutations
			part = &Part{}
			*part = *ap
			part.Header = make(textproto.MIMEHeader)
			root.AddChild(part)
		}
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
