package enmime

import (
	"bufio"
	"encoding/base64"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/textproto"
	"sort"
	"time"

	"github.com/jhillyerd/enmime/internal/coding"
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
	if p.Header == nil {
		p.Header = make(textproto.MIMEHeader)
	}
	cte := p.setupMIMEHeaders()
	// Encode this part.
	b := bufio.NewWriter(writer)
	p.encodeHeader(b)
	if len(p.Content) > 0 {
		b.Write(crnl)
		if err := p.encodeContent(b, cte); err != nil {
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
		b.Write(marker)
		b.Write(crnl)
		if err := c.Encode(b); err != nil {
			return err
		}
		c = c.NextSibling
	}
	b.Write(endMarker)
	b.Write(crnl)
	return b.Flush()
}

// setupMIMEHeaders determines content transfer encoding, generates a boundary string if required,
// then sets the Content-Type (type, charset, filename, boundary) and Content-Disposition headers.
func (p *Part) setupMIMEHeaders() transferEncoding {
	// Determine content transfer encoding.

	// If we are encoding a part that previously had content-transfer-encoding set, unset it so
	// the correct encoding detection can be done below.
	p.Header.Del(hnContentEncoding)

	cte := te7Bit
	if len(p.Content) > 0 {
		cte = teBase64
		if p.TextContent() {
			cte = selectTransferEncoding(p.Content, false)
			if p.Charset == "" {
				p.Charset = utf8
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
		p.Boundary = "enmime-" + stringutil.UUID()
	}
	if p.ContentID != "" {
		p.Header.Set(hnContentID, coding.ToIDHeader(p.ContentID))
	}
	fileName := p.FileName
	switch selectTransferEncoding([]byte(p.FileName), true) {
	case teBase64:
		fileName = mime.BEncoding.Encode(utf8, p.FileName)
	case teQuoted:
		fileName = mime.QEncoding.Encode(utf8, p.FileName)
	}
	if p.ContentType != "" {
		// Build content type header.
		param := make(map[string]string)
		for k, v := range p.ContentTypeParams {
			param[k] = v
		}
		setParamValue(param, hpCharset, p.Charset)
		setParamValue(param, hpName, fileName)
		setParamValue(param, hpBoundary, p.Boundary)
		if mt := mime.FormatMediaType(p.ContentType, param); mt != "" {
			p.ContentType = mt
		}
		p.Header.Set(hnContentType, p.ContentType)
	}
	if p.Disposition != "" {
		// Build disposition header.
		param := make(map[string]string)
		setParamValue(param, hpFilename, fileName)
		if !p.FileModDate.IsZero() {
			setParamValue(param, hpModDate, p.FileModDate.Format(time.RFC822))
		}
		if mt := mime.FormatMediaType(p.Disposition, param); mt != "" {
			p.Disposition = mt
		}
		p.Header.Set(hnContentDisposition, p.Disposition)
	}
	return cte
}

// encodeHeader writes out a sorted list of headers.
func (p *Part) encodeHeader(b *bufio.Writer) {
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
			b.Write(wb)
		}
	}
}

// encodeContent writes out the content in the selected encoding.
func (p *Part) encodeContent(b *bufio.Writer, cte transferEncoding) (err error) {
	switch cte {
	case teBase64:
		enc := base64.StdEncoding
		text := make([]byte, enc.EncodedLen(len(p.Content)))
		base64.StdEncoding.Encode(text, p.Content)
		// Wrap lines.
		lineLen := 76
		for len(text) > 0 {
			if lineLen > len(text) {
				lineLen = len(text)
			}
			if _, err = b.Write(text[:lineLen]); err != nil {
				return err
			}
			b.Write(crnl)
			text = text[lineLen:]
		}
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
	threshold := b64Percent * len(content) / 100
	bincount := 0
	for _, b := range content {
		if (b < ' ' || '~' < b) && b != '\t' {
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

// setParamValue will ignore empty values
func setParamValue(p map[string]string, k, v string) {
	if v != "" {
		p[k] = v
	}
}
