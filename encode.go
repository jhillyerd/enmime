package enmime

import (
	"bufio"
	"encoding/base64"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/textproto"
	"sort"
	"strconv"
	"strings"

	"github.com/jhillyerd/enmime/internal/stringutil"
)

// b64Percent determines the percent of non-ASCII characters enmime will tolerate before switching
// from quoted-printable to base64 encoding.
const b64Percent = 20

type transferEncoding byte

const (
	te7Bit transferEncoding = iota
	teQuoted
	teBase64
)

var crnl = []byte{'\r', '\n'}

// Encode writes this Part and all its children to the specified writer in MIME format.
func (p *Part) Encode(writer io.Writer) error {
	// Determine content transfer encoding.
	cte := te7Bit
	if p.Header == nil {
		p.Header = make(textproto.MIMEHeader)
	}
	if len(p.Content) > 0 {
		cte = teBase64
		if p.TextContent() {
			cte = selectTransferEncoding(p.Content, false)
			if p.Charset == "" {
				p.Charset = utf8 // Default
			}
		}
		// RFC 2045: 7bit is assumed if CTE header not present.
		switch cte {
		case teBase64:
			p.Header.Set(hnContentEncoding, cteBase64)
		case teQuoted:
			p.Header.Set(hnContentEncoding, cteQuotedPrintable)
		}
	}
	// Setup headers.
	if p.FirstChild != nil && p.Boundary == "" {
		// Multipart, generate random boundary marker.
		uuid := stringutil.UUID()
		p.Boundary = "enmime-" + uuid
	}
	if p.ContentID != "" {
		p.Header.Set(hnContentID, p.ContentID)
	}
	if p.ContentType != "" {
		// Build content type header.
		param := make(map[string]string)
		if p.Charset != "" {
			param[hpCharset] = p.Charset
		}
		if p.FileName != "" {
			param[hpName] = quotedString(p.FileName)
		}
		if p.Boundary != "" {
			param[hpBoundary] = p.Boundary
		}
		mt := mime.FormatMediaType(p.ContentType, param)
		if mt == "" {
			// There was an error, FormatMediaType couldn't encode it with the params.
			mt = p.ContentType
		}
		p.Header.Set(hnContentType, mt)
	}
	if p.Disposition != "" {
		// Build disposition header.
		param := make(map[string]string)
		if p.FileName != "" {
			param[hpFilename] = quotedString(p.FileName)
		}
		p.Header.Set(hnContentDisposition, mime.FormatMediaType(p.Disposition, param))
	}
	// Encode this part.
	b := bufio.NewWriter(writer)
	if err := p.encodeHeader(b); err != nil {
		return err
	}
	if len(p.Content) > 0 {
		if _, err := b.Write(crnl); err != nil {
			return err
		}
		if err := p.encodeContent(b, cte); err != nil {
			return err
		}
		if _, err := b.Write(crnl); err != nil {
			return err
		}
	}
	if p.FirstChild == nil {
		return b.Flush()
	}
	// Encode children.
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
	return b.Flush()
}

// encodeHeader writes out a sorted list of headers.
func (p *Part) encodeHeader(b *bufio.Writer) error {
	keys := make([]string, 0, len(p.Header))
	for k := range p.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range p.Header[k] {
			encv := v
			switch selectTransferEncoding([]byte(v), true) {
			case teBase64:
				encv = mime.BEncoding.Encode(utf8, v)
			case teQuoted:
				encv = mime.QEncoding.Encode(utf8, v)
			}
			// _ used to prevent early wrapping
			wb := stringutil.Wrap(76, k, ":_", encv, "\r\n")
			wb[len(k)+1] = ' '
			if _, err := b.Write(wb); err != nil {
				return err
			}
		}
	}
	return nil
}

// encodeContent writes out the content in the selected encoding.
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

// selectTransferEncoding scans content for non-ASCII characters and selects 'b' or 'q' encoding.
func selectTransferEncoding(content []byte, quoteLineBreaks bool) transferEncoding {
	if len(content) == 0 {
		return te7Bit
	}
	// Binary chars remaining before we choose b64 encoding.
	threshold := b64Percent * 100 / len(content)
	bincount := 0
	for _, b := range content {
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

// quotedString escapes non-ASCII characters in s.
func quotedString(s string) string {
	if strings.IndexFunc(s, func(r rune) bool { return r&0x80 != 0 }) >= 0 {
		return strings.Trim(strconv.QuoteToASCII(s), `"`)
	}
	return s
}
