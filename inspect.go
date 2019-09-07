package enmime

import (
	"bufio"
	"bytes"
	"io"
	"mime"
	"net/textproto"
	"strings"

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
			res[h] = append(res[h], rfc2047decode(value))
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

// rfc2047decode returns a decoded string if the input uses RFC2047 encoding, otherwise it will return the input.
// RFC2047 Example:
//     `=?UTF-8?B?bmFtZT0iw7DCn8KUwoo=?=`
func rfc2047decode(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, s)
	var err error
	for {
		s, err = rfc2047recurse(s)
		switch err {
		case nil:
			continue
		default:
			return s
		}
	}
}

// rfc2047recurse is called for if the value contains content encoded in RFC2047 format and decodes it.
// RFC2047 Example:
//     `=?UTF-8?B?bmFtZT0iw7DCn8KUwoo=?=`
func rfc2047recurse(s string) (string, error) {
	us := strings.ToUpper(s)
	if !strings.Contains(us, "?Q?") && !strings.Contains(us, "?B?") {
		return s, io.EOF
	}

	val, err := decodeHeaderWithError(s)
	if err != nil {
		return val, err
	}
	if val == s {
		val, err = decodeHeaderWithError(fixRFC2047String(val))
		if err != nil {
			return val, err
		}
		if val == s {
			return val, io.EOF
		}
	}

	return val, nil
}

// decodeHeaderWithError decodes a single line (per RFC 2047) using Golang's mime.WordDecoder
func decodeHeaderWithError(input string) (string, error) {
	dec := new(mime.WordDecoder)
	dec.CharsetReader = coding.NewCharsetReader
	header, err := dec.DecodeHeader(input)
	if err != nil {
		return input, err
	}
	return header, nil
}

// fixRFC2047String removes the following characters from charset and encoding segments of an RFC2047 string:
//  '\n', '\r' and ' '
func fixRFC2047String(s string) string {
	inString := false
	isWithinTerminatingEqualSigns := false
	questionMarkCount := 0
	sb := &strings.Builder{}
	for _, v := range s {
		switch v {
		case '=':
			if questionMarkCount == 3 {
				inString = false
			} else {
				isWithinTerminatingEqualSigns = true
			}
			sb.WriteRune(v)
		case '?':
			if isWithinTerminatingEqualSigns {
				inString = true
			} else {
				questionMarkCount++
			}
			isWithinTerminatingEqualSigns = false
			sb.WriteRune(v)
		case '\n', '\r', ' ':
			if !inString {
				sb.WriteRune(v)
			}
			isWithinTerminatingEqualSigns = false
		default:
			isWithinTerminatingEqualSigns = false
			sb.WriteRune(v)
		}
	}
	return sb.String()
}
