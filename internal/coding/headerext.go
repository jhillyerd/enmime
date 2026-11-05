package coding

import (
	"fmt"
	"io"
	"mime"
	"regexp"
	"strings"
)

// rfc2231Language matches a language tag as defined by RFC 2231 section 7, which uses the
// RFC 1766 grammar: one or more one to eight letter segments separated by "-", as in "en-US".
var rfc2231Language = regexp.MustCompile(`^[a-zA-Z]{1,8}(-[a-zA-Z]{1,8})*$`)

// NewExtMimeDecoder creates new MIME word decoder which allows decoding of additional charsets.
func NewExtMimeDecoder() *mime.WordDecoder {
	return &mime.WordDecoder{
		CharsetReader: func(charset string, input io.Reader) (io.Reader, error) {
			// RFC 2231 section 5 allows a language after the charset, as in =?UTF-8*en?Q?...?=
			charset, err := stripRFC2231Language(charset)
			if err != nil {
				return nil, err
			}
			return NewCharsetReader(charset, input)
		},
	}
}

// stripRFC2231Language removes the RFC 2231 section 5 language that may follow the charset of
// an encoded word, as in "UTF-8*en".  A malformed language, such as the empty tag in
// "UTF-8*" or "UTF-8*bad*extra", is rejected so the encoded word is left undecoded instead of
// being read with a charset its author never specified.
func stripRFC2231Language(charset string) (string, error) {
	base, language, found := strings.Cut(charset, "*")
	if !found {
		return charset, nil
	}
	if !rfc2231Language.MatchString(language) {
		return "", fmt.Errorf("invalid RFC 2231 language %q in charset %q", language, charset)
	}
	return base, nil
}

// DecodeExtHeader decodes a single line (per RFC 2047, aka Message Header Extensions) using Golang's
// mime.WordDecoder.
func DecodeExtHeader(input string) string {
	if !strings.Contains(input, "=?") {
		// Don't scan if there is nothing to do here
		return input
	}

	header, err := NewExtMimeDecoder().DecodeHeader(input)
	if err != nil {
		return input
	}

	return header
}

// RFC2047Decode returns a decoded string if the input uses RFC2047 encoding, otherwise it will
// return the input.
//
// RFC2047 Example: `=?UTF-8?B?bmFtZT0iw7DCn8KUwoo=?=`
func RFC2047Decode(s string) string {
	// Convert CR/LF to spaces.
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, s)

	var err error
	decoded := false
	for {
		s, err = rfc2047Recurse(s)
		switch err {
		case nil:
			decoded = true
			continue

		default:
			if decoded {
				key, value, found := strings.Cut(s, "=")
				if !found {
					return s
				}

				// Add quotes as needed.
				if !strings.HasPrefix(value, "\"") {
					value = `"` + value
				}
				if !strings.HasSuffix(value, "\"") {
					value += `"`
				}

				return fmt.Sprintf("%s=%s", key, value)
			}

			return s
		}
	}
}

// rfc2047Recurse is called for if the value contains content encoded in RFC2047 format and decodes
// it.
func rfc2047Recurse(s string) (string, error) {
	us := strings.ToUpper(s)
	if !strings.Contains(us, "?Q?") && !strings.Contains(us, "?B?") {
		return s, io.EOF
	}

	var val string
	if val = DecodeExtHeader(s); val == s {
		if val = DecodeExtHeader(fixRFC2047String(val)); val == s {
			return val, io.EOF
		}
	}

	return val, nil
}

// fixRFC2047String removes the following characters from charset and encoding segments of an
// RFC2047 string: '\n', '\r' and ' '
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
