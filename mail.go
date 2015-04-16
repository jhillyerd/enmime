package enmime

import (
	"fmt"
	"mime"
	"net/mail"
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
	}

	return false
}

// ParseMIMEBody parses the body of the message object into a  tree of MIMEPart objects,
// each of which is aware of its content type, filename and headers.  If the part was
// encoded in quoted-printable or base64, it is decoded before being stored in the
// MIMEPart object.
func ParseMIMEBody(mailMsg *mail.Message) (*MIMEBody, error) {
	mimeMsg := &MIMEBody{header: mailMsg.Header}

	if !IsMultipartMessage(mailMsg) {
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
			if mediatype, _, err := mime.ParseMediaType(ctype); err == nil {
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
				mimeMsg.Text = string(match.Content())
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
				mimeMsg.Text += string(m.Content())
			}
		}

		// Locate HTML body
		match := BreadthMatchFirst(root, func(p MIMEPart) bool {
			return p.ContentType() == "text/html" && p.Disposition() != "attachment"
		})
		if match != nil {
			mimeMsg.Html = string(match.Content())
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
	return decodeHeader(m.header.Get(name))
}
