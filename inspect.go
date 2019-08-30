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

func HumanHeadersOnly(b []byte) (map[string]string, error) {
	b = clean(b)
	tr := textproto.NewReader(bufio.NewReader(bytes.NewReader(b)))
	headers, err := tr.ReadMIMEHeader()
	switch errors.Cause(err) {
	case nil, io.EOF:
	// carry on, io.EOF is expected
	default:
		return nil, err
	}
	bs := bufio.NewScanner(bufio.NewReader(bytes.NewReader(b)))
	res := map[string]string{}
	for bs.Scan() {
		line := bs.Text()
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		i := strings.Index(line, ":")
		if i == -1 {
			continue
		}
		header := textproto.CanonicalMIMEHeaderKey(line[:i])
		if len(headers[header]) == 0 {
			// somethings up, we should have already processed all of these, so why are we trying to fetch from an empty list, did we miscount?
			continue
		}
		// pop
		firstValue := headers[header][0]
		// shift
		headers[header] = headers[header][1:]

		res[header] = rfc2047parts(firstValue)
	}

	return res, nil
}

func clean(b []byte) []byte {
	slice := bytes.SplitAfter(b, []byte{'\r', '\n'})
	dest := make([]byte, 0, len(b)+2)
	headers := true
	for _, v := range slice {
		if (bytes.Index(v, []byte{':'}) > -1 || bytes.HasPrefix(v, []byte{' '}) || bytes.HasPrefix(v, []byte{'\t'})) && headers {
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

// rfc2047parts checks if the value contains content encoded in RFC2047 format
// RFC2047 Example:
//     `=?UTF-8?B?bmFtZT0iw7DCn8KUwoo=?=`
func rfc2047parts(s string) string {
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

// rfc2047recurse is called for if the value contains content encoded in RFC2047 format and decodes it
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

func fixRFC2047String(s string) string {
	inString := false
	eq := false
	q := 0
	sb := &strings.Builder{}
	for _, v := range s {
		switch v {
		case '=':
			if q == 3 {
				inString = false
			} else {
				eq = true
			}
			sb.WriteRune(v)
		case '?':
			if eq {
				inString = true
			} else {
				q += 1
			}
			eq = false
			sb.WriteRune(v)
		case '\n', '\r', ' ':
			if !inString {
				sb.WriteRune(v)
			}
			eq = false
		default:
			eq = false
			sb.WriteRune(v)
		}
	}
	return sb.String()
}
