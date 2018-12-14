package enmime

import (
	"bufio"
	"bytes"
	stderrors "errors"
	"io"
	"mime"
	"net/textproto"
	"strings"

	"github.com/jhillyerd/enmime/internal/coding"
	"github.com/pkg/errors"
)

const (
	// Standard MIME content dispositions
	cdAttachment = "attachment"
	cdInline     = "inline"

	// Standard MIME content types
	ctAppOctetStream   = "application/octet-stream"
	ctMultipartAltern  = "multipart/alternative"
	ctMultipartMixed   = "multipart/mixed"
	ctMultipartPrefix  = "multipart/"
	ctMultipartRelated = "multipart/related"
	ctTextPlain        = "text/plain"
	ctTextHTML         = "text/html"

	// Used as a placeholder in case of malformed Content-Type headers
	ctPlaceholder = "x-not-a-mime-type/x-not-a-mime-type"
	// Used as a placeholder param value in case of malformed
	// Content-Type/Content-Disposition parameters that lack values.
	// E.g.: Content-Type: text/html;iso-8859-1
	pvPlaceholder = "not-a-param-value"

	// Standard Transfer encodings
	cte7Bit            = "7bit"
	cte8Bit            = "8bit"
	cteBase64          = "base64"
	cteBinary          = "binary"
	cteQuotedPrintable = "quoted-printable"

	// Standard MIME header names
	hnContentDisposition = "Content-Disposition"
	hnContentEncoding    = "Content-Transfer-Encoding"
	hnContentID          = "Content-ID"
	hnContentType        = "Content-Type"
	hnMIMEVersion        = "MIME-Version"

	// Standard MIME header parameters
	hpBoundary = "boundary"
	hpCharset  = "charset"
	hpFile     = "file"
	hpFilename = "filename"
	hpName     = "name"
	hpModDate  = "modification-date"

	utf8 = "utf-8"
)

var errEmptyHeaderBlock = stderrors.New("empty header block")

// AddressHeaders is the set of SMTP headers that contain email addresses, used by
// Envelope.AddressList().  Key characters must be all lowercase.
var AddressHeaders = map[string]bool{
	"bcc":             true,
	"cc":              true,
	"delivered-to":    true,
	"from":            true,
	"reply-to":        true,
	"to":              true,
	"sender":          true,
	"resent-bcc":      true,
	"resent-cc":       true,
	"resent-from":     true,
	"resent-reply-to": true,
	"resent-to":       true,
	"resent-sender":   true,
}

// Terminology from RFC 2047:
//  encoded-word: the entire =?charset?encoding?encoded-text?= string
//  charset: the character set portion of the encoded word
//  encoding: the character encoding type used for the encoded-text
//  encoded-text: the text we are decoding

// readHeader reads a block of SMTP or MIME headers and returns a textproto.MIMEHeader.
// Header parse warnings & errors will be added to p.Errors, io errors will be returned directly.
func readHeader(r *bufio.Reader, p *Part) (textproto.MIMEHeader, error) {
	// buf holds the massaged output for textproto.Reader.ReadMIMEHeader()
	buf := &bytes.Buffer{}
	tp := textproto.NewReader(r)
	firstHeader := true
	for {
		// Pull out each line of the headers as a temporary slice s
		s, err := tp.ReadLineBytes()
		if err != nil {
			cause := errors.Cause(err)
			if cause == io.ErrUnexpectedEOF && buf.Len() == 0 {
				return nil, errors.WithStack(errEmptyHeaderBlock)
			} else if cause == io.EOF {
				buf.Write([]byte{'\r', '\n'})
				break
			}
			return nil, err
		}
		firstColon := bytes.IndexByte(s, ':')
		firstSpace := bytes.IndexAny(s, " \t\n\r")
		if firstSpace == 0 {
			// Starts with space: continuation
			buf.WriteByte(' ')
			buf.Write(textproto.TrimBytes(s))
			continue
		}
		if firstColon == 0 {
			// Can't parse line starting with colon: skip
			p.addError(ErrorMalformedHeader, "Header line %q started with a colon", s)
			continue
		}
		if firstColon > 0 {
			// Contains a colon, treat as a new header line
			if !firstHeader {
				// New Header line, end the previous
				buf.Write([]byte{'\r', '\n'})
			}
			s = textproto.TrimBytes(s)
			buf.Write(s)
			firstHeader = false
		} else {
			// No colon: potential non-indented continuation
			if len(s) > 0 {
				// Attempt to detect and repair a non-indented continuation of previous line
				buf.WriteByte(' ')
				buf.Write(s)
				p.addWarning(ErrorMalformedHeader, "Continued line %q was not indented", s)
			} else {
				// Empty line, finish header parsing
				buf.Write([]byte{'\r', '\n'})
				break
			}
		}
	}
	buf.Write([]byte{'\r', '\n'})
	tr := textproto.NewReader(bufio.NewReader(buf))
	header, err := tr.ReadMIMEHeader()
	return header, errors.WithStack(err)
}

// decodeHeader decodes a single line (per RFC 2047) using Golang's mime.WordDecoder
func decodeHeader(input string) string {
	if !strings.Contains(input, "=?") {
		// Don't scan if there is nothing to do here
		return input
	}

	dec := new(mime.WordDecoder)
	dec.CharsetReader = coding.NewCharsetReader
	header, err := dec.DecodeHeader(input)
	if err != nil {
		return input
	}
	return header
}

// decodeToUTF8Base64Header decodes a MIME header per RFC 2047, reencoding to =?utf-8b?
func decodeToUTF8Base64Header(input string) string {
	if !strings.Contains(input, "=?") {
		// Don't scan if there is nothing to do here
		return input
	}

	tokens := strings.FieldsFunc(input, whiteSpaceRune)
	output := make([]string, len(tokens))
	for i, token := range tokens {
		if len(token) > 4 && strings.Contains(token, "=?") {
			// Stash parenthesis, they should not be encoded
			prefix := ""
			suffix := ""
			if token[0] == '(' {
				prefix = "("
				token = token[1:]
			}
			if token[len(token)-1] == ')' {
				suffix = ")"
				token = token[:len(token)-1]
			}
			// Base64 encode token
			output[i] = prefix + mime.BEncoding.Encode("UTF-8", decodeHeader(token)) + suffix
		} else {
			output[i] = token
		}
	}

	// Return space separated tokens
	return strings.Join(output, " ")
}

// parseMediaType is a more tolerant implementation of Go's mime.ParseMediaType function.
func parseMediaType(ctype string) (mtype string, params map[string]string, invalidParams []string, err error) {
	mtype, params, err = mime.ParseMediaType(ctype)
	if err != nil {
		// Small hack to remove harmless charset duplicate params.
		mctype := fixMangledMediaType(ctype, ";")
		mtype, params, err = mime.ParseMediaType(mctype)
		if err != nil {
			// Some badly formed media types forget to send ; between fields.
			mctype := fixMangledMediaType(ctype, " ")
			if strings.Contains(mctype, `name=""`) {
				mctype = strings.Replace(mctype, `name=""`, `name=" "`, -1)
			}
			mtype, params, err = mime.ParseMediaType(mctype)
			if err != nil {
				// If the media parameter has special characters, ensure that it is quoted.
				mtype, params, err = mime.ParseMediaType(fixUnquotedSpecials(mctype))
				if err != nil {
					return "", nil, nil, errors.WithStack(err)
				}
			}
		}
	}
	if mtype == ctPlaceholder {
		mtype = ""
	}
	for name, value := range params {
		if value != pvPlaceholder {
			continue
		}
		invalidParams = append(invalidParams, name)
		delete(params, name)
	}
	return mtype, params, invalidParams, err
}

// fixMangledMediaType is used to insert ; separators into media type strings that lack them, and
// remove repeated parameters.
func fixMangledMediaType(mtype, sep string) string {
	if mtype == "" {
		return ""
	}
	parts := strings.Split(mtype, sep)
	mtype = ""
	for i, p := range parts {
		switch i {
		case 0:
			if p == "" {
				// The content type is completely missing. Put in a placeholder.
				p = ctPlaceholder
			}
		default:
			if !strings.Contains(p, "=") {
				p = p + "=" + pvPlaceholder
			}

			// RFC-2047 encoded attribute name
			p = rfc2047AttributeName(p)

			pair := strings.Split(p, "=")
			if strings.Contains(mtype, pair[0]+"=") {
				// Ignore repeated parameters.
				continue
			}

			if strings.ContainsAny(pair[0], "()<>@,;:\"\\/[]?") {
				// attribute is a strict token and cannot be a quoted-string
				// if any of the above characters are present in a token it
				// must be quoted and is therefor an invalid attribute.
				// Discard the pair.
				continue
			}
		}
		mtype += p
		// Only terminate with semicolon if not the last parameter and if it doesn't already have a
		// semicolon.
		if i != len(parts)-1 && !strings.HasSuffix(mtype, ";") {
			mtype += ";"
		}
	}
	if strings.HasSuffix(mtype, ";") {
		mtype = mtype[:len(mtype)-1]
	}
	return mtype
}

// consumeParam takes the the parameter part of a Content-Type header, returns a clean version of
// the first parameter (quoted as necessary), and the remainder of the parameter part of the
// Content-Type header.
//
// Given this this header:
//     `Content-Type: text/calendar; charset=utf-8; method=text/calendar`
// `consumeParams` should be given this part:
//     ` charset=utf-8; method=text/calendar`
// And returns (first pass):
//     `consumed = "charset=utf-8;"`
//     `rest     = " method=text/calendar"`
// Capture the `consumed` value (to build a clean Content-Type header value) and pass the value of
// `rest` back to `consumeParam`. That second call will return:
//     `consumed = " method=\"text/calendar\""`
//     `rest     = ""`
// Again, use the value of `consumed` to build a clean Content-Type header value. Given that `rest`
// is empty, all of the parameters have been consumed successfully.
//
// If `consumed` is returned empty and `rest` is not empty, then the value of `rest` does not
// begin with a parsable parameter. This does not necessarily indicate a problem. For example,
// if there is trailing whitespace, it would be returned here.
func consumeParam(s string) (consumed, rest string) {
	i := strings.IndexByte(s, '=')
	if i < 0 {
		return "", s
	}

	param := strings.Builder{}
	param.WriteString(s[:i+1])
	s = s[i+1:]

	value := strings.Builder{}
	valueQuotedOriginally := false
	valueQuoteAdded := false
	valueQuoteNeeded := false

	var r rune
findValueStart:
	for i, r = range s {
		switch r {
		case ' ', '\t':
			param.WriteRune(r)

		case '"':
			valueQuotedOriginally = true
			valueQuoteAdded = true
			value.WriteRune(r)

			break findValueStart

		default:
			valueQuotedOriginally = false
			valueQuoteAdded = false
			value.WriteRune(r)

			break findValueStart
		}
	}

	if len(s)-i < 1 {
		// parameter value starts at the end of the string, make empty
		// quoted string to play nice with mime.ParseMediaType
		param.WriteString(`""`)

	} else {
		// The beginning of the value is not at the end of the string

		s = s[i+1:]

		quoteIfUnquoted := func() {
			if !valueQuoteNeeded {
				if !valueQuoteAdded {
					param.WriteByte('"')

					valueQuoteAdded = true
				}

				valueQuoteNeeded = true
			}
		}

	findValueEnd:
		for len(s) > 0 {
			switch s[0] {
			case ';', ' ', '\t':
				if valueQuotedOriginally {
					// We're in a quoted string, so whitespace is allowed.
					value.WriteByte(s[0])
					s = s[1:]
					break
				}

				// Otherwise, we've reached the end of an unquoted value.

				param.WriteString(value.String())
				value.Reset()

				if valueQuoteNeeded {
					param.WriteByte('"')
				}

				param.WriteByte(s[0])
				s = s[1:]

				break findValueEnd

			case '"':
				if valueQuotedOriginally {
					// We're in a quoted value. This is the end of that value.
					param.WriteString(value.String())
					value.Reset()

					param.WriteByte(s[0])
					s = s[1:]

					break findValueEnd
				}

				quoteIfUnquoted()

				value.WriteByte('\\')
				value.WriteByte(s[0])
				s = s[1:]

			case '\\':
				if len(s) > 1 {
					value.WriteByte(s[0])
					s = s[1:]

					// Backslash escapes the next char. Consume that next char.
					value.WriteByte(s[0])

					quoteIfUnquoted()
				}
				// Else there is no next char to consume.
				s = s[1:]

			case '(', ')', '<', '>', '@', ',', ':', '/', '[', ']', '?', '=':
				quoteIfUnquoted()

				fallthrough

			default:
				value.WriteByte(s[0])
				s = s[1:]
			}
		}
	}

	if value.Len() > 0 {
		// There is a value that ends with the string. Capture it.
		param.WriteString(value.String())

		if valueQuotedOriginally || valueQuoteNeeded {
			// If valueQuotedOriginally is true and we got here,
			// that means there was no closing quote. So we'll add one.
			// Otherwise, we're here because it was an unquoted value
			// with a special char in it, and we had to quote it.
			param.WriteByte('"')
		}
	}

	return param.String(), s
}

// fixUnquotedSpecials as defined in RFC 2045, section 5.1:
// https://tools.ietf.org/html/rfc2045#section-5.1
func fixUnquotedSpecials(s string) string {
	idx := strings.IndexByte(s, ';')
	if idx < 0 || idx == len(s) {
		// No parameters
		return s
	}

	clean := strings.Builder{}
	clean.WriteString(s[:idx+1])
	s = s[idx+1:]

	for len(s) > 0 {
		var consumed string
		consumed, s = consumeParam(s)

		if len(consumed) == 0 {
			clean.WriteString(s)
			return clean.String()
		}

		clean.WriteString(consumed)
	}

	return clean.String()
}

// Detects a RFC-822 linear-white-space, passed to strings.FieldsFunc.
func whiteSpaceRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

// rfc2047AttributeName checks if the attribute name is encoded in RFC2047 format
// RFC2047 Example:
//     `=?UTF-8?B?bmFtZT0iw7DCn8KUwoo=?=`
func rfc2047AttributeName(s string) string {
	if !strings.Contains(s, "?=") {
		return s
	}
	pair := strings.SplitAfter(s, "?=")
	pair[0] = decodeHeader(pair[0])
	return strings.Join(pair, "")
}
