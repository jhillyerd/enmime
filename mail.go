package enmime

import (
	"fmt"
	"mime"
	"net/mail"
)

type MIMEBody struct {
	Text string
	Html string
	Root *MIMEPart
}

func IsMIMEMessage(mailMsg *mail.Message) bool {
	// Parse top-level multipart
	ctype := mailMsg.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ctype)
	if err != nil {
		return false
	}
	switch mediatype {
	case "multipart/alternative":
		return true
	}

	return false
}

func ParseMIMEBody(mailMsg *mail.Message) (*MIMEBody, error) {
	mimeMsg := new(MIMEBody)

	if !IsMIMEMessage(mailMsg) {
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
		switch mediatype {
		case "multipart/alternative":
			// Good
		default:
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
		match := root.BreadthFirstSearch(func(node *MIMEPart) bool {
			return node.Type == "text/plain"
		})
		if match != nil {
			mimeMsg.Text = string(match.Content)
		}

		// Locate HTML body
		match = root.BreadthFirstSearch(func(node *MIMEPart) bool {
			return node.Type == "text/html"
		})
		if match != nil {
			mimeMsg.Html = string(match.Content)
		}
	}

	return mimeMsg, nil
}

func (m *MIMEBody) String() string {
	return fmt.Sprintf("----TEXT----\n%v\n----HTML----\n%v\n----END----\n", m.Text, m.Html)
}
