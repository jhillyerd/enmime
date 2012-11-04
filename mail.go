package enmime

import (
	"fmt"
	"mime"
	"net/mail"
	"strings"
)

type MIMEBody struct {
	Text        string
	Html        string
	Root        MIMEPart
	Attachments []MIMEPart
	Inlines     []MIMEPart
}

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
		"multipart/related":
		return true
	}

	return false
}

func ParseMIMEBody(mailMsg *mail.Message) (*MIMEBody, error) {
	mimeMsg := new(MIMEBody)

	if !IsMultipartMessage(mailMsg) {
		// Parse as text only
		bodyBytes, err := decodeSection(mailMsg.Header.Get("Content-Transfer-Encoding"), mailMsg.Body)
		if err != nil {
			return nil, err
		}
		mimeMsg.Text = string(bodyBytes)
	} else {
		// Parse top-level multipart
		ctype := mailMsg.Header.Get("Content-Type")
		mediatype, params, err := mime.ParseMediaType(ctype)
		if err != nil {
			return nil, err
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
		err = parseParts(root, mailMsg.Body, boundary)
		if err != nil {
			return nil, err
		}

		// Locate text body
		match := BreadthMatchFirst(root, func(p MIMEPart) bool {
			return p.ContentType() == "text/plain" && p.Disposition() != "attachment"
		})
		if match != nil {
			mimeMsg.Text = string(match.Content())
		}

		// Locate HTML body
		match = BreadthMatchFirst(root, func(p MIMEPart) bool {
			return p.ContentType() == "text/html" && p.Disposition() != "attachment"
		})
		if match != nil {
			mimeMsg.Html = string(match.Content())
		}

		// Locate attachments
		mimeMsg.Attachments = BreadthMatchAll(root, func(p MIMEPart) bool {
			return p.Disposition() == "attachment"
		})

		// Locate inlines
		mimeMsg.Inlines = BreadthMatchAll(root, func(p MIMEPart) bool {
			return p.Disposition() == "inline"
		})
	}

	return mimeMsg, nil
}

func (m *MIMEBody) String() string {
	return fmt.Sprintf("----TEXT----\n%v\n----HTML----\n%v\n----END----\n", m.Text, m.Html)
}
