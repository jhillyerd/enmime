package enmime

import (
	"fmt"
	"mime"
	"net/mail"
	"net/textproto"
	"strings"
)

// MIMEBody is the outer wrapper for MIME messages.
type MIMEBody struct {
	Text        string      // The plain text portion of the message
	Html        string      // The HTML portion of the message
	Root        MIMEPart    // The top-level MIMEPart
	Attachments []MIMEPart  // All parts having a Content-Disposition of attachment
	Inlines     []MIMEPart  // All parts having a Content-Disposition of inline
	OtherParts  []MIMEPart  // All parts not in Attachments and Inlines
	header      mail.Header // Header from original message
}

var AddressHeaders = []string{"From", "To", "Delivered-To", "Cc", "Bcc", "Reply-To"}

// IsMultipartMessage returns true if the message has a recognized multipart Content-Type
// header.  You don't need to check this before calling ParseMIMEBody, it can handle
// non-multipart messages.
func IsMultipartMessage(mailMsg *mail.Message) bool {
	// Parse top-level multipart
	ctype := mailMsg.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ctype)
	if err != nil {
		return false
	}
	switch mediatype {
	case "multipart/alternative",
		"multipart/mixed",
		"multipart/related",
		"multipart/signed":
		return true
	default:
		if strings.HasPrefix(mediatype, "multipart/") {
			// according to rfc2046#section-5.1.7 all other multipart should
			// be treated as multipart/mixed
			return true
		}
	}
	return false
}

// IsAttachment returns true, if the given header defines an attachment.
// First it checks, if the Content-Disposition header defines an attachement.
// If this test is false, the Content-Type header is checked.
//
// Valid Attachment-Headers:
//
//    Content-Disposition: attachment; filename="frog.jpg"
//    Content-Type: attachment; filename="frog.jpg"
//
func IsAttachment(header mail.Header) bool {
	mediatype, _, _ := mime.ParseMediaType(header.Get("Content-Disposition"))
	if strings.ToLower(mediatype) == "attachment" {
		return true
	}

	mediatype, _, _ = mime.ParseMediaType(header.Get("Content-Type"))
	if strings.ToLower(mediatype) == "attachment" {
		return true
	}

	return false
}

// IsPlain returns true, if the the mime headers define a valid
// 'text/plain' or 'text/html part'. Ff emptyContentTypeIsPlain is set
// to true, a missing Content-Type header will result in a positive
// plain part detection.
func IsPlain(header mail.Header, emptyContentTypeIsPlain bool) bool {
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

// IsBinaryBody returns true, if the mail header defines a binary body.
func IsBinaryBody(mailMsg *mail.Message) bool {
	if IsAttachment(mailMsg.Header) == true {
		return true
	}

	return !IsPlain(mailMsg.Header, true)
}

// Returns a MIME message with only one Attachment, the parsed original mail body.
func binMIME(mailMsg *mail.Message) (*MIMEBody, error) {
	// Root Node of our tree
	ctype := mailMsg.Header.Get("Content-Type")
	mediatype, mparams, err := mime.ParseMediaType(ctype)
	if err != nil {
		mediatype = "attachment"
	}

	m := &MIMEBody{
		header: mailMsg.Header,
		Root:   NewMIMEPart(nil, mediatype),
	}

	p := NewMIMEPart(nil, mediatype)
	p.content, err = decodeSection(mailMsg.Header.Get("Content-Transfer-Encoding"), mailMsg.Body)
	if err != nil {
		return nil, err
	}

	// get set headers
	p.header = make(textproto.MIMEHeader, 4)
	// Figure out our disposition, filename
	disposition, dparams, err := mime.ParseMediaType(mailMsg.Header.Get("Content-Disposition"))
	if err == nil {
		// Disposition is optional
		p.disposition = disposition
		p.fileName = DecodeHeader(dparams["filename"])
	}
	if p.fileName == "" && mparams["name"] != "" {
		p.fileName = DecodeHeader(mparams["name"])
	}
	if p.fileName == "" && mparams["file"] != "" {
		p.fileName = DecodeHeader(mparams["file"])
	}
	if p.charset == "" {
		p.charset = mparams["charset"]
	}

	p.header.Set("Content-Type", mailMsg.Header.Get("Content-Type"))
	p.header.Set("Content-Disposition", mailMsg.Header.Get("Content-Disposition"))

	m.Attachments = append(m.Attachments, p)
	return m, err
}

// ParseMIMEBody parses the body of the message object into a  tree of MIMEPart objects,
// each of which is aware of its content type, filename and headers.  If the part was
// encoded in quoted-printable or base64, it is decoded before being stored in the
// MIMEPart object.
func ParseMIMEBody(mailMsg *mail.Message) (*MIMEBody, error) {
	mimeMsg := &MIMEBody{header: mailMsg.Header}

	if !IsMultipartMessage(mailMsg) {

		// Attachment only?
		if IsBinaryBody(mailMsg) {
			return binMIME(mailMsg)
		}

		// Parse as text only
		bodyBytes, err := decodeSection(mailMsg.Header.Get("Content-Transfer-Encoding"),
			mailMsg.Body)
		if err != nil {
			return nil, fmt.Errorf("Error decoding text-only message: %v", err)
		}
		mimeMsg.Text = string(bodyBytes)

		// Check for HTML at top-level, eat errors quietly
		ctype := mailMsg.Header.Get("Content-Type")
		if ctype != "" {
			if mediatype, mparams, err := mime.ParseMediaType(ctype); err == nil {
				/*
				 *Content-Type: text/plain;\t charset="hz-gb-2312"
				 */
				if mparams["charset"] != "" {
					newStr, err := ConvertToUTF8String(mparams["charset"], bodyBytes)
					if err != nil {
						return nil, err
					}
					mimeMsg.Text = newStr
				} else if mediatype == "text/html" { // charset is empty, look in html body for charset
					charset, err := charsetFromHTMLString(mimeMsg.Text)

					if charset != "" && err == nil {
						newStr, err := ConvertToUTF8String(charset, bodyBytes)
						if err == nil {
							mimeMsg.Text = newStr
						}
					}
				}
				if mediatype == "text/html" {
					mimeMsg.Html = mimeMsg.Text
				}
			}
		}
	} else {
		// Parse top-level multipart
		ctype := mailMsg.Header.Get("Content-Type")
		mediatype, params, err := mime.ParseMediaType(ctype)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse media type: %v", err)
		}
		if !strings.HasPrefix(mediatype, "multipart/") {
			return nil, fmt.Errorf("Unknown mediatype: %v", mediatype)
		}
		boundary := params["boundary"]
		if boundary == "" {
			return nil, fmt.Errorf("Unable to locate boundary param in Content-Type header")
		}
		// Root Node of our tree
		root := NewMIMEPart(nil, mediatype)
		mimeMsg.Root = root
		err = parseParts(root, mailMsg.Body, boundary)
		if err != nil {
			return nil, err
		}

		// Locate text body
		if mediatype == "multipart/altern" {
			match := BreadthMatchFirst(root, func(p MIMEPart) bool {
				return p.ContentType() == "text/plain" && p.Disposition() != "attachment"
			})
			if match != nil {
				if match.Charset() != "" {
					newStr, err := ConvertToUTF8String(match.Charset(), match.Content())
					if err != nil {
						return nil, err
					}
					mimeMsg.Text += newStr
				} else {
					mimeMsg.Text += string(match.Content())
				}
			}
		} else {
			// multipart is of a mixed type
			match := DepthMatchAll(root, func(p MIMEPart) bool {
				return p.ContentType() == "text/plain" && p.Disposition() != "attachment"
			})
			for i, m := range match {
				if i > 0 {
					mimeMsg.Text += "\n--\n"
				}
				if m.Charset() != "" {
					newStr, err := ConvertToUTF8String(m.Charset(), m.Content())
					if err != nil {
						return nil, err
					}
					mimeMsg.Text += newStr
				} else {
					mimeMsg.Text += string(m.Content())
				}
			}
		}

		// Locate HTML body
		match := BreadthMatchFirst(root, func(p MIMEPart) bool {
			return p.ContentType() == "text/html" && p.Disposition() != "attachment"
		})
		if match != nil {
			if match.Charset() != "" {
				newStr, err := ConvertToUTF8String(match.Charset(), match.Content())
				if err != nil {
					return nil, err
				}
				mimeMsg.Html += newStr
			} else {
				mimeMsg.Html = string(match.Content())
			}

		}

		// Locate attachments
		mimeMsg.Attachments = BreadthMatchAll(root, func(p MIMEPart) bool {
			return p.Disposition() == "attachment" || p.ContentType() == "application/octet-stream"
		})

		// Locate inlines
		mimeMsg.Inlines = BreadthMatchAll(root, func(p MIMEPart) bool {
			return p.Disposition() == "inline"
		})

		// Locate others parts not handled in "Attachments" and "inlines"
		mimeMsg.OtherParts = BreadthMatchAll(root, func(p MIMEPart) bool {
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
	}

	return mimeMsg, nil
}

// Process the specified header for RFC 2047 encoded words and return the result
func (m *MIMEBody) GetHeader(name string) string {
	return DecodeHeader(m.header.Get(name))
}

// Return AddressList with RFC 2047 encoded encoded names.
func (m *MIMEBody) AddressList(key string) ([]*mail.Address, error) {
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

	str := DecodeToUTF8Base64Header(m.header.Get(key))
	if str == "" {
		return nil, mail.ErrHeaderNotPresent
	}
	ret, err := mail.ParseAddressList(str)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
