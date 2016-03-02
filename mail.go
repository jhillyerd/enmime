package enmime

import (
	"fmt"
	"github.com/jaytaylor/html2text"
	"mime"
	"net/mail"
	"net/textproto"
	"strings"
)

// MIMEBody is the outer wrapper for MIME messages.
type MIMEBody struct {
	Text           string      // The plain text portion of the message
	HTML           string      // The HTML portion of the message
	IsTextFromHTML bool        // Plain text was empty; down-converted HTML
	Root           MIMEPart    // The top-level MIMEPart
	Attachments    []MIMEPart  // All parts having a Content-Disposition of attachment
	Inlines        []MIMEPart  // All parts having a Content-Disposition of inline
	OtherParts     []MIMEPart  // All parts not in Attachments and Inlines
	header         mail.Header // Header from original message
}

// AddressHeaders enumerates SMTP headers that contain email addresses
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

// IsAttachment returns true, if the given header defines an attachment.  First
// it checks if the Content-Disposition header defines an attachement or inline
// attachment. If this test is false, the Content-Type header is checked for
// attachment, but not inline.  Email clients use inline for their text bodies.
//
// Valid Attachment-Headers:
//
//    Content-Disposition: attachment; filename="frog.jpg"
//    Content-Disposition: inline; filename="frog.jpg"
//    Content-Type: attachment; filename="frog.jpg"
//
func IsAttachment(header mail.Header) bool {
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

// IsPlain returns true, if the the MIME headers define a valid 'text/plain' or
// 'text/html' part. If the emptyContentTypeIsPlain argument is set to true, a
// missing Content-Type header will result in a positive plain part detection.
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

// IsBinaryBody returns true if the mail header defines a binary body.
func IsBinaryBody(mailMsg *mail.Message) bool {
	if IsAttachment(mailMsg.Header) == true {
		return true
	}

	return !IsPlain(mailMsg.Header, true)
}

// binMIME handles the special case where the only content of the message is an
// attachment.  It is called by ParseMIME when needed.
func binMIME(mailMsg *mail.Message) (*MIMEBody, error) {
	// Determine mediatype
	ctype := mailMsg.Header.Get("Content-Type")
	mediatype, mparams, err := mime.ParseMediaType(ctype)
	if err != nil {
		mediatype = "attachment"
	}

	// Build the MIME part representing most of this message
	p := NewMIMEPart(nil, mediatype)
	content, err := decodeSection(mailMsg.Header.Get("Content-Transfer-Encoding"), mailMsg.Body)
	if err != nil {
		return nil, err
	}
	p.SetContent(content)
	p.SetHeader(make(textproto.MIMEHeader, 4))

	// Determine and set headers for: content disposition, filename and
	// character set
	disposition, dparams, err := mime.ParseMediaType(mailMsg.Header.Get("Content-Disposition"))
	if err == nil {
		// Disposition is optional
		p.SetDisposition(disposition)
		p.SetFileName(DecodeHeader(dparams["filename"]))
	}
	if p.FileName() == "" && mparams["name"] != "" {
		p.SetFileName(DecodeHeader(mparams["name"]))
	}
	if p.FileName() == "" && mparams["file"] != "" {
		p.SetFileName(DecodeHeader(mparams["file"]))
	}
	if p.Charset() == "" {
		p.SetCharset(mparams["charset"])
	}

	p.Header().Set("Content-Type", mailMsg.Header.Get("Content-Type"))
	p.Header().Set("Content-Disposition", mailMsg.Header.Get("Content-Disposition"))

	// Add our part to the appropriate section of MIMEBody
	m := &MIMEBody{
		header:         mailMsg.Header,
		Root:           NewMIMEPart(nil, mediatype),
		IsTextFromHTML: false,
	}

	if disposition == "inline" {
		m.Inlines = append(m.Inlines, p)
	} else {
		m.Attachments = append(m.Attachments, p)
	}

	return m, err
}

// ParseMIMEBody parses the body of the message object into a  tree of MIMEPart
// objects, each of which is aware of its content type, filename and headers.
// If the part was encoded in quoted-printable or base64, it is decoded before
// being stored in the MIMEPart object.
func ParseMIMEBody(mailMsg *mail.Message) (*MIMEBody, error) {
	mimeMsg := &MIMEBody{
		IsTextFromHTML: false,
		header:         mailMsg.Header,
	}

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
		// Handle plain ASCII text, content-type unspecified
		mimeMsg.Text = string(bodyBytes)

		// Process top-level content-type
		ctype := mailMsg.Header.Get("Content-Type")
		if ctype != "" {
			if mediatype, mparams, err := mime.ParseMediaType(ctype); err == nil {
				if mparams["charset"] != "" {
					// Convert plain text to UTF8 if content type specified a charset
					newStr, err := ConvertToUTF8String(mparams["charset"], bodyBytes)
					if err != nil {
						return nil, err
					}
					mimeMsg.Text = newStr
				} else if mediatype == "text/html" {
					// charset is empty, look in HTML body for charset
					charset, err := charsetFromHTMLString(mimeMsg.Text)

					if charset != "" && err == nil {
						newStr, err := ConvertToUTF8String(charset, bodyBytes)
						if err == nil {
							mimeMsg.Text = newStr
						}
					}
				}
				if mediatype == "text/html" {
					mimeMsg.HTML = mimeMsg.Text
					// Empty Text will trigger html2text conversion below
					mimeMsg.Text = ""
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
				mimeMsg.HTML += newStr
			} else {
				mimeMsg.HTML = string(match.Content())
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

// GetHeader processes the specified header for RFC 2047 encoded words and
// return the result
func (m *MIMEBody) GetHeader(name string) string {
	return DecodeHeader(m.header.Get(name))
}

// AddressList returns a mail.Address slice with RFC 2047 encoded encoded names.
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
	// These statements are handy for debugging ParseAddressList errors
	// fmt.Println("in:  ", m.header.Get(key))
	// fmt.Println("out: ", str)
	ret, err := mail.ParseAddressList(str)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
