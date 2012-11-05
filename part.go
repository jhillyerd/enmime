package enmime

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"github.com/sloonz/go-qprintable"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
)

type MIMEPart interface {
	Parent() MIMEPart
	FirstChild() MIMEPart
	NextSibling() MIMEPart
	Header() textproto.MIMEHeader
	ContentType() string
	Disposition() string
	FileName() string
	// TODO Content should probably be a reader
	Content() []byte
}

// memMIMEPart contains a single part of a multipart MIME document in memory,
// the tree may be navigated via the Parent, FirstChild and NextSibling pointers.
type memMIMEPart struct {
	parent      MIMEPart
	firstChild  MIMEPart
	nextSibling MIMEPart
	header      textproto.MIMEHeader
	contentType string
	disposition string
	fileName    string
	content     []byte
}

func NewMIMEPart(parent MIMEPart, contentType string) *memMIMEPart {
	return &memMIMEPart{parent: parent, contentType: contentType}
}

func (p *memMIMEPart) Parent() MIMEPart {
	return p.parent
}

func (p *memMIMEPart) FirstChild() MIMEPart {
	return p.firstChild
}

func (p *memMIMEPart) NextSibling() MIMEPart {
	return p.nextSibling
}

func (p *memMIMEPart) Header() textproto.MIMEHeader {
	return p.header
}

func (p *memMIMEPart) ContentType() string {
	return p.contentType
}

func (p *memMIMEPart) Disposition() string {
	return p.disposition
}

func (p *memMIMEPart) FileName() string {
	return p.fileName
}

func (p *memMIMEPart) Content() []byte {
	return p.content
}

// ParseMIME reads a MIME document from the provided reader and parses it into
// tree of MIMEPart objects.
func ParseMIME(reader *bufio.Reader) (MIMEPart, error) {
	tr := textproto.NewReader(reader)
	header, err := tr.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	mediatype, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	root := &memMIMEPart{header: header, contentType: mediatype}

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
		root.content = content
	}

	return root, nil
}

// parseParts recursively parses a mime multipart document.
func parseParts(parent *memMIMEPart, reader io.Reader, boundary string) error {
	var prevSibling *memMIMEPart

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
		mediatype, mparams, err := mime.ParseMediaType(mrp.Header.Get("Content-Type"))
		if err != nil {
			return err
		}

		// Insert ourselves into tree, p is go-mime's mime-part
		p := NewMIMEPart(parent, mediatype)
		if prevSibling != nil {
			prevSibling.nextSibling = p
		} else {
			parent.firstChild = p
		}
		prevSibling = p

		// Figure out our disposition, filename
		disposition, dparams, err := mime.ParseMediaType(mrp.Header.Get("Content-Disposition"))
		if err == nil {
			// Disposition is optional
			p.disposition = disposition
			p.fileName = dparams["filename"]
		}
		if p.fileName == "" && mparams["name"] != "" {
			p.fileName = mparams["name"]
		}

		boundary := mparams["boundary"]
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
			p.content = data
		}
	}

	return nil
}

// decodeSection attempts to decode the data from reader using the algorithm listed in
// the Content-Transfer-Encoding header, returning the raw data if it does not known
// the encoding type.
func decodeSection(encoding string, reader io.Reader) ([]byte, error) {
	// Default is to just read input into bytes
	decoder := reader

	switch strings.ToLower(encoding) {
	case "quoted-printable":
		decoder = qprintable.NewDecoder(qprintable.WindowsTextEncoding, reader)
	case "base64":
		cleaner := NewBase64Cleaner(reader)
		decoder = base64.NewDecoder(base64.StdEncoding, cleaner)
	}
	
	// Read bytes into buffer
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(decoder)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}


