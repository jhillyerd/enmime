package enmime

import (
	"bytes"
	"errors"
	"io/ioutil"
	"mime"
	"net/mail"
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
func Builder() MailBuilder {
	return MailBuilder{}
}

// Error returns the stored error from a file attachment/inline read or nil.
func (p MailBuilder) Error() error {
	return p.err
}

// Date returns a copy of MailBuilder with the specified Date header.
func (p MailBuilder) Date(date time.Time) MailBuilder {
	p.date = date
	return p
}

// From returns a copy of MailBuilder with the specified From header.
func (p MailBuilder) From(name, addr string) MailBuilder {
	p.from = mail.Address{Name: name, Address: addr}
	return p
}

// Subject returns a copy of MailBuilder with the specified Subject header.
func (p MailBuilder) Subject(subject string) MailBuilder {
	p.subject = subject
	return p
}

// To returns a copy of MailBuilder with this name & address appended to the To header.  name may be
// empty.
func (p MailBuilder) To(name, addr string) MailBuilder {
	if len(addr) > 0 {
		p.to = append(p.to, mail.Address{Name: name, Address: addr})
	}
	return p
}

// ToAddrs returns a copy of MailBuilder with the specified To addresses.
func (p MailBuilder) ToAddrs(to []mail.Address) MailBuilder {
	p.to = to
	return p
}

// CC returns a copy of MailBuilder with this name & address appended to the CC header.  name may be
// empty.
func (p MailBuilder) CC(name, addr string) MailBuilder {
	if len(addr) > 0 {
		p.cc = append(p.cc, mail.Address{Name: name, Address: addr})
	}
	return p
}

// CCAddrs returns a copy of MailBuilder with the specified CC addresses.
func (p MailBuilder) CCAddrs(cc []mail.Address) MailBuilder {
	p.cc = cc
	return p
}

// BCC returns a copy of MailBuilder with this name & address appended to the BCC list.  name may be
// empty.  This method only has an effect if the Send method is used to transmit the message, there
// is no effect on the parts returned by Build().
func (p MailBuilder) BCC(name, addr string) MailBuilder {
	if len(addr) > 0 {
		p.bcc = append(p.bcc, mail.Address{Name: name, Address: addr})
	}
	return p
}

// BCCAddrs returns a copy of MailBuilder with the specified as the blind CC list.  This method only
// has an effect if the Send method is used to transmit the message, there is no effect on the parts
// returned by Build().
func (p MailBuilder) BCCAddrs(bcc []mail.Address) MailBuilder {
	p.bcc = bcc
	return p
}

// ReplyTo returns a copy of MailBuilder with this name & address appended to the To header.  name
// may be empty.
func (p MailBuilder) ReplyTo(name, addr string) MailBuilder {
	p.replyTo = mail.Address{Name: name, Address: addr}
	return p
}

// Header returns a copy of MailBuilder with the specified value added to the named header.
func (p MailBuilder) Header(name, value string) MailBuilder {
	// Copy existing header map
	h := textproto.MIMEHeader{}
	for k, v := range p.header {
		h[k] = v
	}
	h.Add(name, value)
	p.header = h
	return p
}

// Text returns a copy of MailBuilder that will use the provided bytes for its text/plain Part.
func (p MailBuilder) Text(body []byte) MailBuilder {
	p.text = body
	return p
}

// HTML returns a copy of MailBuilder that will use the provided bytes for its text/html Part.
func (p MailBuilder) HTML(body []byte) MailBuilder {
	p.html = body
	return p
}

// AddAttachment returns a copy of MailBuilder that includes the specified attachment.
func (p MailBuilder) AddAttachment(b []byte, contentType string, fileName string) MailBuilder {
	part := NewPart(contentType)
	part.Content = b
	part.FileName = fileName
	part.Disposition = cdAttachment
	p.attachments = append(p.attachments, part)
	return p
}

// AddFileAttachment returns a copy of MailBuilder that includes the specified attachment.
// fileName, will be populated from the base name of path.  Content type will be detected from the
// path extension.
func (p MailBuilder) AddFileAttachment(path string) MailBuilder {
	// Only allow first p.err value
	if p.err != nil {
		return p
	}
	f, err := os.Open(path)
	if err != nil {
		p.err = err
		return p
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		p.err = err
		return p
	}
	name := filepath.Base(path)
	ctype := mime.TypeByExtension(filepath.Ext(name))
	return p.AddAttachment(b, ctype, name)
}

// AddInline returns a copy of MailBuilder that includes the specified inline.  fileName and
// contentID may be left empty.
func (p MailBuilder) AddInline(
	b []byte,
	contentType string,
	fileName string,
	contentID string,
) MailBuilder {
	part := NewPart(contentType)
	part.Content = b
	part.FileName = fileName
	part.Disposition = cdInline
	part.ContentID = contentID
	p.inlines = append(p.inlines, part)
	return p
}

// AddFileInline returns a copy of MailBuilder that includes the specified inline.  fileName and
// contentID will be populated from the base name of path.  Content type will be detected from the
// path extension.
func (p MailBuilder) AddFileInline(path string) MailBuilder {
	// Only allow first p.err value
	if p.err != nil {
		return p
	}
	f, err := os.Open(path)
	if err != nil {
		p.err = err
		return p
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		p.err = err
		return p
	}
	name := filepath.Base(path)
	ctype := mime.TypeByExtension(filepath.Ext(name))
	return p.AddInline(b, ctype, name, name)
}

// Build performs some basic validations, then constructs a tree of Part structs from the configured
// MailBuilder.  It will set the Date header to now if it was not explicitly set.
func (p MailBuilder) Build() (*Part, error) {
	if p.err != nil {
		return nil, p.err
	}
	// Validations
	if p.from.Address == "" {
		return nil, errors.New("from not set")
	}
	if len(p.to)+len(p.cc)+len(p.bcc) == 0 {
		return nil, errors.New(ErrorMissingRecipient)
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
		root = NewPart(ctTextPlain)
		root.Content = p.text
		root.Charset = utf8
	}
	if p.html != nil {
		part = NewPart(ctTextHTML)
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
		root = NewPart(ctMultipartAltern)
		root.AddChild(part)
	}
	if len(p.inlines) > 0 {
		part = root
		root = NewPart(ctMultipartRelated)
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
		root = NewPart(ctMultipartMixed)
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

// SendWithReversePath encodes the message and sends it via the specified Sender.
func (p MailBuilder) SendWithReversePath(sender Sender, from string) error {
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
	return sender.Send(from, recips, buf.Bytes())
}

// Send encodes the message and sends it via the specified Sender, using the address provided to
// `From()` as the reverse-path.
func (p MailBuilder) Send(sender Sender) error {
	return p.SendWithReversePath(sender, p.from.Address)
}

// Equals uses the reflect package to test two MailBuilder structs for equality, primarily for unit
// tests.
func (p MailBuilder) Equals(o MailBuilder) bool {
	return reflect.DeepEqual(p, o)
}
