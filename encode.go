package enmime

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/textproto"
	"sort"

	"github.com/jhillyerd/enmime/internal/stringutil"
)

// b64Percent determines the percent of non-ASCII characters enmime will tolerate before switching
// from quoted-printable to base64 encoding
const b64Percent = 20

type transferEncoding byte

const (
	te7Bit transferEncoding = iota
	teQuoted
	teBase64
)

var crnl = []byte{'\r', '\n'}

// Encode writes this Part and all its children to the specified writer in MIME format
func (p *Part) Encode(writer io.Writer) error {
	// Determine content transfer encoding
	cte := te7Bit
	if len(p.Content) > 0 {
		cte = teBase64
		if p.TextContent() {
			cte = selectTransferEncoding(string(p.Content), false)
			if p.Charset == "" {
				p.Charset = "utf-8" // Default
			}
		}
		// RFC 2045: 7bit is assumed if CTE header not present
		switch cte {
		case teBase64:
			p.Header.Set(hnContentEncoding, cteBase64)
		case teQuoted:
			p.Header.Set(hnContentEncoding, cteQuotedPrintable)
		}
	}
	// Setup headers
	if p.Header == nil {
		p.Header = make(textproto.MIMEHeader)
	}
	if p.FirstChild != nil && p.Boundary == "" {
		// Multipart, generate random boundary marker
		uuid, err := newUUID()
		if err != nil {
			return err
		}
		p.Boundary = "enmime-boundary-" + uuid
	}
	if p.ContentType != "" {
		// Build content type header
		param := make(map[string]string)
		if p.Charset != "" {
			param[hpCharset] = p.Charset
		}
		if p.FileName != "" {
			param[hpName] = p.FileName
		}
		if p.Boundary != "" {
			param[hpBoundary] = p.Boundary
		}
		p.Header.Set(hnContentType, mime.FormatMediaType(p.ContentType, param))
	}
	if p.Disposition != "" {
		// Build disposition header
		param := make(map[string]string)
		if p.FileName != "" {
			param[hpFilename] = p.FileName
		}
		p.Header.Set(hnContentDisposition, mime.FormatMediaType(p.Disposition, param))
	}
	// Encode this part
	b := bufio.NewWriter(writer)
	if err := p.encodeHeader(b); err != nil {
		return err
	}
	if err := p.encodeContent(b, cte); err != nil {
		return err
	}
	if _, err := b.Write(crnl); err != nil {
		return err
	}
	if p.FirstChild != nil {
		// Encode children
		endMarker := []byte("\r\n--" + p.Boundary + "--")
		marker := endMarker[:len(endMarker)-2]
		c := p.FirstChild
		for c != nil {
			if _, err := b.Write(marker); err != nil {
				return err
			}
			if _, err := b.Write(crnl); err != nil {
				return err
			}
			if err := c.Encode(b); err != nil {
				return err
			}
			c = c.NextSibling
		}
		if _, err := b.Write(endMarker); err != nil {
			return err
		}
		if _, err := b.Write(crnl); err != nil {
			return err
		}

	}
	return b.Flush()
}

// encodeHeader writes out a sorted list of headers
func (p *Part) encodeHeader(b *bufio.Writer) error {
	keys := make([]string, 0, len(p.Header))
	for k := range p.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range p.Header[k] {
			encv := v
			switch selectTransferEncoding(v, true) {
			case teBase64:
				encv = mime.BEncoding.Encode("utf-8", v)
			case teQuoted:
				encv = mime.QEncoding.Encode("utf-8", v)
			}
			// _ used to prevent early wrapping
			wb := stringutil.Wrap(76, k, ":_", encv, "\r\n")
			wb[len(k)+1] = ' '
			if _, err := b.Write(wb); err != nil {
				return err
			}
		}
	}
	_, err := b.Write([]byte{'\r', '\n'})
	return err
}

// encodeContent writes out the content in the selected encoding
func (p *Part) encodeContent(b *bufio.Writer, cte transferEncoding) (err error) {
	switch cte {
	case teBase64:
		enc := base64.StdEncoding
		text := make([]byte, enc.EncodedLen(len(p.Content)))
		base64.StdEncoding.Encode(text, p.Content)
		lineLen := 76
		for len(text) > 0 {
			// Wrap lines
			if lineLen > len(text) {
				lineLen = len(text)
			}
			if _, err = b.Write(text[:lineLen]); err != nil {
				return err
			}
			if _, err = b.Write(crnl); err != nil {
				return err
			}
			text = text[lineLen:]
		}
		_, err = b.Write(text)
	case teQuoted:
		qp := quotedprintable.NewWriter(b)
		if _, err = qp.Write(p.Content); err != nil {
			return err
		}
		err = qp.Close()
	default:
		_, err = b.Write(p.Content)
	}
	return err
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]),
		nil
}

// selectTransferEncoding scans the string for non-ASCII characters and selects 'b' or 'q' encoding
func selectTransferEncoding(s string, quoteLineBreaks bool) transferEncoding {
	// binary chars remaining before we choose b64 encoding
	threshold := b64Percent * 100 / len(s)
	bincount := 0
	for _, b := range s {
		if (b < ' ' || b > '~') && b != '\t' {
			if !quoteLineBreaks && (b == '\r' || b == '\n') {
				continue
			}
			bincount++
			if bincount >= threshold {
				return teBase64
			}
		}
	}
	if bincount == 0 {
		return te7Bit
	}
	return teQuoted
}
