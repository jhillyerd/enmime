package enmime

import (
	"bytes"
	"errors"
	"io/ioutil"
	"mime"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/jhillyerd/enmime/internal/stringutil"
)

// MailBuilder facilitates the easy construction of a MIME message.  Each manipulation method
// returns a copy of the receiver struct.  It can be considered immutable if the caller does not
// modify the string and byte slices passed in.  Immutability allows the headers or entire message
// to be reused across multiple threads.
type MailBuilder struct {
	to, cc, bcc          []mail.Address
	from                 mail.Address
	replyTo              mail.Address
	subject              string
	date                 time.Time
	header               textproto.MIMEHeader
	text, html           []byte
	inlines, attachments []*Part
	err                  error
}

// Builder returns an empty MailBuilder struct.
func Builder() *MailBuilder {
	return &MailBuilder{}
}

// Error returns the stored error from a file attachment/inline read or nil.
func (p *MailBuilder) Error() error {
	return p.err
}

// Date returns a copy of MailBuilder with the specified Date header.
func (p *MailBuilder) Date(date time.Time) *MailBuilder {
	c := *p
	c.date = date
	return &c
}

// From returns a copy of MailBuilder with the specified From header.
func (p *MailBuilder) From(name, addr string) *MailBuilder {
	c := *p
	c.from = mail.Address{Name: name, Address: addr}
	return &c
}

// Subject returns a copy of MailBuilder with the specified Subject header.
func (p *MailBuilder) Subject(subject string) *MailBuilder {
	c := *p
	c.subject = subject
	return &c
}

// To returns a copy of MailBuilder with this name & address appended to the To header.  name may be
// empty.
func (p *MailBuilder) To(name, addr string) *MailBuilder {
	c := *p
	c.to = append(p.to, mail.Address{Name: name, Address: addr})
	return &c
}

// ToAddrs returns a copy of MailBuilder with the specified To addresses.
func (p *MailBuilder) ToAddrs(to []mail.Address) *MailBuilder {
	c := *p
	c.to = to
	return &c
}

// CC returns a copy of MailBuilder with this name & address appended to the CC header.  name may be
// empty.
func (p *MailBuilder) CC(name, addr string) *MailBuilder {
	c := *p
	c.cc = append(p.cc, mail.Address{Name: name, Address: addr})
	return &c
}

// CCAddrs returns a copy of MailBuilder with the specified CC addresses.
func (p *MailBuilder) CCAddrs(cc []mail.Address) *MailBuilder {
	c := *p
	c.cc = cc
	return &c
}

// BCC returns a copy of MailBuilder with this name & address appended to the BCC list.  name may be
// empty.  This method only has an effect if the Send method is used to transmit the message, there
// is no effect on the parts returned by Build().
func (p *MailBuilder) BCC(name, addr string) *MailBuilder {
	c := *p
	c.bcc = append(p.bcc, mail.Address{Name: name, Address: addr})
	return &c
}

// BCCAddrs returns a copy of MailBuilder with the specified as the blind CC list.  This method only
// has an effect if the Send method is used to transmit the message, there is no effect on the parts
// returned by Build().
func (p *MailBuilder) BCCAddrs(bcc []mail.Address) *MailBuilder {
	c := *p
	c.bcc = bcc
	return &c
}

// ReplyTo returns a copy of MailBuilder with this name & address appended to the To header.  name
// may be empty.
func (p *MailBuilder) ReplyTo(name, addr string) *MailBuilder {
	c := *p
	c.replyTo = mail.Address{Name: name, Address: addr}
	return &c
}

// Header returns a copy of MailBuilder with the specified value added to the named header.
func (p *MailBuilder) Header(name, value string) *MailBuilder {
	c := *p
	// Copy existing header map
	h := textproto.MIMEHeader{}
	for k, v := range p.header {
		h[k] = v
	}
	h.Add(name, value)
	c.header = h
	return &c
}

// Text returns a copy of MailBuilder that will use the provided bytes for its text/plain Part.
func (p *MailBuilder) Text(body []byte) *MailBuilder {
	c := *p
	c.text = body
	return &c
}

// HTML returns a copy of MailBuilder that will use the provided bytes for its text/html Part.
func (p *MailBuilder) HTML(body []byte) *MailBuilder {
	c := *p
	c.html = body
	return &c
}

// AddAttachment returns a copy of MailBuilder that includes the specified attachment.
func (p *MailBuilder) AddAttachment(b []byte, contentType string, fileName string) *MailBuilder {
	part := NewPart(nil, contentType)
	part.Content = b
	part.FileName = fileName
	part.Disposition = cdAttachment
	c := *p
	c.attachments = append(c.attachments, part)
	return &c
}

// AddFileAttachment returns a copy of MailBuilder that includes the specified attachment.
// fileName, will be populated from the base name of path.  Content type will be detected from the
// path extension.
func (p *MailBuilder) AddFileAttachment(path string) *MailBuilder {
	c := *p
	// Only allow first p.err value
	if c.err != nil {
		return &c
	}
	f, err := os.Open(path)
	if err != nil {
		c.err = err
		return &c
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		c.err = err
		return &c
	}
	name := filepath.Base(path)
	ctype := mime.TypeByExtension(filepath.Ext(name))
	return p.AddAttachment(b, ctype, name)
}

// AddInline returns a copy of MailBuilder that includes the specified inline.  fileName and
// contentID may be left empty.
func (p *MailBuilder) AddInline(
	b []byte,
	contentType string,
	fileName string,
	contentID string,
) *MailBuilder {
	part := NewPart(nil, contentType)
	part.Content = b
	part.FileName = fileName
	part.Disposition = cdInline
	part.ContentID = contentID
	c := *p
	c.inlines = append(c.inlines, part)
	return &c
}

// AddFileInline returns a copy of MailBuilder that includes the specified inline.  fileName and
// contentID will be populated from the base name of path.  Content type will be detected from the
// path extension.
func (p *MailBuilder) AddFileInline(path string) *MailBuilder {
	c := *p
	// Only allow first p.err value
	if c.err != nil {
		return &c
	}
	f, err := os.Open(path)
	if err != nil {
		c.err = err
		return &c
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		c.err = err
		return &c
	}
	name := filepath.Base(path)
	ctype := mime.TypeByExtension(filepath.Ext(name))
	return p.AddInline(b, ctype, name, name)
}

// Build performs some basic validations, then constructs a tree of Part structs from the configured
// MailBuilder.  It will set the Date header to now if it was not explicitly set.
func (p *MailBuilder) Build() (*Part, error) {
	if p.err != nil {
		return nil, p.err
	}
	// Validations
	if p.from.Address == "" {
		return nil, errors.New("from not set")
	}
	if p.subject == "" {
		return nil, errors.New("subject not set")
	}
	if len(p.to)+len(p.cc)+len(p.bcc) == 0 {
		return nil, errors.New("no recipients (to, cc, bcc) set")
	}
	// Fully loaded structure; the presence of text, html, inlines, and attachments will determine
	// how much is necessary:
	//
	//  multipart/mixed
	//  |- multipart/related
	//  |  |- multipart/alternative
	//  |  |  |- text/plain
	//  |  |  `- text/html
	//  |  `- inlines..
	//  `- attachments..
	//
	// We build this tree starting at the leaves, re-rooting as needed.
	var root, part *Part
	if p.text != nil || p.html == nil {
		root = NewPart(nil, ctTextPlain)
		root.Content = p.text
		root.Charset = utf8
	}
	if p.html != nil {
		part = NewPart(nil, ctTextHTML)
		part.Content = p.html
		part.Charset = utf8
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
	if len(p.inlines) > 0 {
		part = root
		root = NewPart(nil, ctMultipartRelated)
		root.AddChild(part)
		for _, ip := range p.inlines {
			// Copy inline Part to isolate mutations
			part = &Part{}
			*part = *ip
			part.Header = make(textproto.MIMEHeader)
			root.AddChild(part)
		}
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
	h.Set(hnMIMEVersion, "1.0")
	h.Set("From", p.from.String())
	h.Set("Subject", p.subject)
	if len(p.to) > 0 {
		h.Set("To", stringutil.JoinAddress(p.to))
	}
	if len(p.cc) > 0 {
		h.Set("Cc", stringutil.JoinAddress(p.cc))
	}
	if p.replyTo.Address != "" {
		h.Set("Reply-To", p.replyTo.String())
	}
	date := p.date
	if date.IsZero() {
		date = time.Now()
	}
	h.Set("Date", date.Format(time.RFC1123Z))
	for k, v := range p.header {
		for _, s := range v {
			h.Add(k, s)
		}
	}
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
	for _, a := range p.to {
		recips = append(recips, a.Address)
	}
	for _, a := range p.cc {
		recips = append(recips, a.Address)
	}
	for _, a := range p.bcc {
		recips = append(recips, a.Address)
	}
	return smtp.SendMail(addr, a, p.from.Address, recips, buf.Bytes())
}

// Equals uses the reflect package to test two MailBuilder structs for equality, primarily for unit
// tests.
func (p *MailBuilder) Equals(o *MailBuilder) bool {
	if p == nil {
		return o == nil
	}
	return reflect.DeepEqual(p, o)
}
