package enmime

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"mime"
	"sort"

	"github.com/jhillyerd/enmime/internal/stringutil"
)

// b64Percent determines the percent of non-ASCII characters enmime will tolerate in a header before
// switching from quoted-printable to base64 encoding
const b64Percent = 20

var crnl = []byte{'\r', '\n'}

// Encode writes this Part and all its children to the specified writer in MIME format
func (p *Part) Encode(writer io.Writer) error {
	b := bufio.NewWriter(writer)
	// Encode this part
	if err := p.encodeHeader(b); err != nil {
		return err
	}
	if _, err := b.Write(p.Content); err != nil {
		return err
	}
	if _, err := b.Write(crnl); err != nil {
		return err
	}
	if p.FirstChild != nil {
		// Encode children
		if p.Boundary == "" {
			// Generate random boundary marker
			uuid, err := newUUID()
			if err != nil {
				return err
			}
			p.Boundary = "enmime-boundary-" + uuid
		}
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
			we := selectEncoder(v)
			// _ used to prevent early wrapping
			wb := stringutil.Wrap(76, k, ":_", we.Encode("utf-8", v), "\r\n")
			wb[len(k)+1] = ' '
			if _, err := b.Write(wb); err != nil {
				return err
			}
		}
	}
	_, err := b.Write([]byte{'\r', '\n'})
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

// selectEncoder scans the string for non-ASCII characters and selects 'b' or 'q' encoding
func selectEncoder(s string) mime.WordEncoder {
	// binary chars remaining before we choose b64 encoding
	binrem := b64Percent * 100 / len(s)
	for _, b := range s {
		if (b < ' ' || b > '~') && b != '\t' {
			binrem--
			if binrem <= 0 {
				return mime.BEncoding
			}
		}
	}
	return mime.QEncoding
}
