package enmime

import (
	"bufio"
	"bytes"
	"io"
	"net/textproto"

	"github.com/jhillyerd/enmime/internal/coding"
	"github.com/pkg/errors"
)

var defaultHeadersList = []string{
	"From",
	"To",
	"Sender",
	"CC",
	"BCC",
	"Subject",
	"Date",
}

// DecodeHeaders returns a limited selection of mime headers for use by user agents
// Default header list:
//   "Date", "Subject", "Sender", "From", "To", "CC" and "BCC"
//
// Additional headers provided will be formatted canonically:
//   h, err := enmime.DecodeHeaders(b, "content-type", "user-agent")
func DecodeHeaders(b []byte, addtlHeaders ...string) (textproto.MIMEHeader, error) {
	b = ensureHeaderBoundary(b)
	tr := textproto.NewReader(bufio.NewReader(bytes.NewReader(b)))
	headers, err := tr.ReadMIMEHeader()
	switch errors.Cause(err) {
	case nil, io.EOF:
	// carry on, io.EOF is expected
	default:
		return nil, err
	}
	headerList := defaultHeadersList
	headerList = append(headerList, addtlHeaders...)
	res := map[string][]string{}
	for _, header := range headerList {
		h := textproto.CanonicalMIMEHeaderKey(header)
		res[h] = make([]string, 0, len(headers[h]))
		for _, value := range headers[h] {
			res[h] = append(res[h], coding.RFC2047Decode(value))
		}
	}

	return res, nil
}

// ensureHeaderBoundary scans through an rfc822 document to ensure the boundary between headers and body exists
func ensureHeaderBoundary(b []byte) []byte {
	slice := bytes.SplitAfter(b, []byte{'\r', '\n'})
	dest := make([]byte, 0, len(b)+2)
	headers := true
	for _, v := range slice {
		if headers && (bytes.Contains(v, []byte{':'}) || bytes.HasPrefix(v, []byte{' '}) || bytes.HasPrefix(v, []byte{'\t'})) {
			dest = append(dest, v...)
			continue
		}
		if headers {
			headers = false
			if !bytes.Equal(v, []byte{'\r', '\n'}) {
				dest = append(dest, append([]byte{'\r', '\n'}, v...)...)
				continue
			}
		}
		dest = append(dest, v...)
	}

	return dest
}
