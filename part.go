package enmime

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/sloonz/go-qprintable"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
)

type MIMEPart struct {
	Parent      *MIMEPart
	FirstChild  *MIMEPart
	NextSibling *MIMEPart
	Type        string
	Header      textproto.MIMEHeader
	Content     []byte
}

func NewMIMEPart(parent *MIMEPart, contentType string) *MIMEPart {
	return &MIMEPart{Parent: parent, Type: contentType}
}

func (n *MIMEPart) String() string {
	children := ""
	siblings := ""
	if n.FirstChild != nil {
		children = n.FirstChild.String()
	}
	if n.NextSibling != nil {
		siblings = n.NextSibling.String()
	}
	return fmt.Sprintf("[%v %v] %v", n.Type, children, siblings)
}

// decodeSection attempts to decode the data from reader using the algorithm listed in
// the Content-Transfer-Encoding header, returning the raw data if it does not known
// the encoding type.
func decodeSection(encoding string, reader io.Reader) ([]byte, error) {
	switch strings.ToLower(encoding) {
	case "quoted-printable":
		decoder := qprintable.NewDecoder(qprintable.WindowsTextEncoding, reader)
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(decoder)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	// Don't know this type, just return bytes
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func parseRoot(reader *bufio.Reader) (*MIMEPart, error) {
	tr := textproto.NewReader(reader)
	header, err := tr.ReadMIMEHeader()
	if err != nil {
	  return nil, err
	}
	mediatype, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
	  return nil, err
	}
	root := &MIMEPart{Header: header, Type: mediatype}

	if strings.HasPrefix(mediatype, "multipart/") {
		boundary := params["boundary"]
		err = parseParts(root, reader, boundary)
		if err != nil {
			return nil, err
		}
	} else {
		// Content is text or data, decode it
		content, err := decodeSection(header.Get("Content-Transfer-Encoding"), reader)
		if err != nil {
			return nil, err
		}
		root.Content = content
	}

	return root, nil
}

func parseParts(parent *MIMEPart, reader io.Reader, boundary string) error {
	var prevSibling *MIMEPart

	// Loop over MIME parts
	mr := multipart.NewReader(reader, boundary)
	for {
		// mrp is go's build in mime-part
		mrp, err := mr.NextPart()
		if err != nil {
			if err == io.EOF {
				// This is a clean end-of-message signal
				break
			}
			return err
		}
		mediatype, params, err := mime.ParseMediaType(mrp.Header.Get("Content-Type"))
		if err != nil {
			return err
		}

		// Insert ourselves into tree, p is go-mime's mime-part
		p := NewMIMEPart(parent, mediatype)
		if prevSibling != nil {
			prevSibling.NextSibling = p
		} else {
			parent.FirstChild = p
		}
		prevSibling = p

		boundary := params["boundary"]
		if boundary != "" {
			// Content is another multipart
			err = parseParts(p, mrp, boundary)
			if err != nil {
				return err
			}
		} else {
			// Content is text or data, decode it
			data, err := decodeSection(mrp.Header.Get("Content-Transfer-Encoding"), mrp)
			if err != nil {
				return err
			}
			p.Content = data
		}
	}

	return nil
}
