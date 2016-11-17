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

// AddressHeaders enumerates SMTP headers that contain email addresses, used by
// Envelope.AddressList()
var AddressHeaders = []string{"From", "To", "Delivered-To", "Cc", "Bcc", "Reply-To"}

// Envelope is a simplified wrapper for MIME email messages.
type Envelope struct {
	Text           string      // The plain text portion of the message
	HTML           string      // The HTML portion of the message
	IsTextFromHTML bool        // Plain text was empty; down-converted HTML
	Root           *Part       // The top-level Part
	Attachments    []*Part     // All parts having a Content-Disposition of attachment
	Inlines        []*Part     // All parts having a Content-Disposition of inline
	OtherParts     []*Part     // All parts not in Attachments and Inlines
	Errors         []*Error    // Errors encountered while parsing
	header         mail.Header // Header from original message
}

// GetHeader processes the specified header for RFC 2047 encoded words and returns the result as a
// UTF-8 string
func (m *Envelope) GetHeader(name string) string {
	return decodeHeader(m.header.Get(name))
}

// AddressList returns a mail.Address slice with RFC 2047 encoded names converted to UTF-8
func (m *Envelope) AddressList(key string) ([]*mail.Address, error) {
	isAddrHeader := false
	for _, hkey := range AddressHeaders {
		if strings.ToLower(hkey) == strings.ToLower(key) {
			isAddrHeader = true
			break
		}
	}
	if !isAddrHeader {
		return nil, fmt.Errorf("%s is not address header", key)
	}

	str := decodeToUTF8Base64Header(m.header.Get(key))
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

// EnvelopeFromMessage parses the body of the mailMsg into an Envelope, downconverting HTML to plain
// text if needed, and sorting the attachments, inlines and other parts into their respective
// slices.
func EnvelopeFromMessage(mailMsg *mail.Message) (*Envelope, error) {
	mimeMsg := &Envelope{
		IsTextFromHTML: false,
		header:         mailMsg.Header,
	}

	if isMultipartMessage(mailMsg) {
		// Multi-part message (message with attachments, etc)
		if err := parseMultiPartBody(mailMsg, mimeMsg); err != nil {
			return nil, err
		}
	} else {
		if isBinaryBody(mailMsg) {
			// Attachment only, no text
			if err := parseBinaryOnlyBody(mailMsg, mimeMsg); err != nil {
				return nil, err
			}
		}
		// Only text, no attachments
		if err := parseTextOnlyBody(mailMsg, mimeMsg); err != nil {
			return nil, err
		}
	}

	// Copy part errors into mimeMsg
	if mimeMsg.Root != nil {
		_ = mimeMsg.Root.DepthMatchAll(func(part *Part) bool {
			// Using DepthMatchAll to traverse all parts, don't care about result
			for _, perr := range part.errors {
				mimeMsg.Errors = append(mimeMsg.Errors, &perr)
			}
			return false
		})
	}

	// Down-convert HTML to text if necessary
	if mimeMsg.Text == "" && mimeMsg.HTML != "" {
		mimeMsg.IsTextFromHTML = true
		var err error
		if mimeMsg.Text, err = html2text.FromString(mimeMsg.HTML); err != nil {
			// Fail gently
			mimeMsg.Text = ""
			return mimeMsg, err
		}
	}

	return mimeMsg, nil
}

// parseTextOnlyBody parses a plain text message in mailMsg that has MIME-like headers, but
// only contains a single part - no boundaries, etc.  The result is placed in mimeMsg.
func parseTextOnlyBody(mailMsg *mail.Message, mimeMsg *Envelope) error {
	bodyBytes, err := decodeSection(
		mailMsg.Header.Get("Content-Transfer-Encoding"), mailMsg.Body)
	if err != nil {
		return fmt.Errorf("Error decoding text-only message: %v", err)
	}

	// Handle plain ASCII text, content-type unspecified, may be reverted later
	mimeMsg.Text = string(bodyBytes)

	// Process top-level content-type
	ctype := mailMsg.Header.Get("Content-Type")
	if ctype != "" {
		if mediatype, mparams, err := mime.ParseMediaType(ctype); err == nil {
			if mparams["charset"] != "" {
				// Convert plain text to UTF8 if content type specified a charset
				newStr, err := convertToUTF8String(mparams["charset"], bodyBytes)
				if err != nil {
					return err
				}
				mimeMsg.Text = newStr
			} else if mediatype == "text/html" {
				// charset is empty, look in HTML body for charset
				charset, err := charsetFromHTMLString(mimeMsg.Text)
				if charset != "" && err == nil {
					newStr, err := convertToUTF8String(charset, bodyBytes)
					if err == nil {
						mimeMsg.Text = newStr
					}
				}
			}
			if mediatype == "text/html" {
				mimeMsg.HTML = mimeMsg.Text
				// Empty Text should trigger html2text conversion
				mimeMsg.Text = ""
			}
		}
	}

	return nil
}

// parseBinaryOnlyBody parses a message where the only content is a binary attachment with no
// other parts. The result is placed in mimeMsg.
func parseBinaryOnlyBody(mailMsg *mail.Message, mimeMsg *Envelope) error {
	// Determine mediatype
	ctype := mailMsg.Header.Get("Content-Type")
	mediatype, mparams, err := mime.ParseMediaType(ctype)
	if err != nil {
		mediatype = "attachment"
	}

	// Build the MIME part representing most of this message
	p := NewPart(nil, mediatype)
	content, err := decodeSection(mailMsg.Header.Get("Content-Transfer-Encoding"), mailMsg.Body)
	if err != nil {
		return err
	}
	p.SetContent(content)
	p.Header = make(textproto.MIMEHeader, 4)

	// Determine and set headers for: content disposition, filename and character set
	disposition, dparams, err := mime.ParseMediaType(mailMsg.Header.Get("Content-Disposition"))
	if err == nil {
		// Disposition is optional
		p.SetDisposition(disposition)
		p.SetFileName(decodeHeader(dparams["filename"]))
	}
	if p.FileName() == "" && mparams["name"] != "" {
		p.SetFileName(decodeHeader(mparams["name"]))
	}
	if p.FileName() == "" && mparams["file"] != "" {
		p.SetFileName(decodeHeader(mparams["file"]))
	}
	if p.Charset() == "" {
		p.SetCharset(mparams["charset"])
	}

	p.Header.Set("Content-Type", mailMsg.Header.Get("Content-Type"))
	p.Header.Set("Content-Disposition", mailMsg.Header.Get("Content-Disposition"))

	// Add our part to the appropriate section of the Envelope
	mimeMsg.Root = NewPart(nil, mediatype)

	if disposition == "inline" {
		mimeMsg.Inlines = append(mimeMsg.Inlines, p)
	} else {
		mimeMsg.Attachments = append(mimeMsg.Attachments, p)
	}

	return nil
}

// parseMultiPartBody parses a multipart message in mailMsg.  The result is placed in mimeMsg.
func parseMultiPartBody(mailMsg *mail.Message, mimeMsg *Envelope) error {
	// Parse top-level multipart
	ctype := mailMsg.Header.Get("Content-Type")
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
	// Root Node of our tree
	root := NewPart(nil, mediatype)
	mimeMsg.Root = root
	err = parseParts(root, mailMsg.Body, boundary)
	if err != nil {
		return err
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
			mimeMsg.Text += string(allBytes)
		}
	} else {
		// multipart is of a mixed type
		match := root.DepthMatchAll(func(p *Part) bool {
			return p.ContentType() == "text/plain" && p.Disposition() != "attachment"
		})
		for i, m := range match {
			if i > 0 {
				mimeMsg.Text += "\n--\n"
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
			mimeMsg.Text += string(allBytes)

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
		mimeMsg.HTML += string(allBytes)
	}

	// Locate attachments
	mimeMsg.Attachments = root.BreadthMatchAll(func(p *Part) bool {
		return p.Disposition() == "attachment" || p.ContentType() == "application/octet-stream"
	})

	// Locate inlines
	mimeMsg.Inlines = root.BreadthMatchAll(func(p *Part) bool {
		return p.Disposition() == "inline"
	})

	// Locate others parts not considered in attachments or inlines
	mimeMsg.OtherParts = root.BreadthMatchAll(func(p *Part) bool {
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
func isMultipartMessage(mailMsg *mail.Message) bool {
	// Parse top-level multipart
	ctype := mailMsg.Header.Get("Content-Type")
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
func isAttachment(header mail.Header) bool {
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
func isPlain(header mail.Header, emptyContentTypeIsPlain bool) bool {
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
func isBinaryBody(mailMsg *mail.Message) bool {
	if isAttachment(mailMsg.Header) == true {
		return true
	}

	return !isPlain(mailMsg.Header, true)
}
