package enmime

import (
	"bufio"
	"bytes"
	"fmt"
	"mime"
	"net/textproto"
	"regexp"
	"strings"

	"github.com/jhillyerd/enmime/internal/coding"
	"github.com/pkg/errors"
)

const (
	// Standard MIME content dispositions
	cdAttachment = "attachment"
	cdInline     = "inline"

	// Standard MIME content types
	ctAppPrefix        = "application/"
	ctAppOctetStream   = "application/octet-stream"
	ctMultipartAltern  = "multipart/alternative"
	ctMultipartMixed   = "multipart/mixed"
	ctMultipartPrefix  = "multipart/"
	ctMultipartRelated = "multipart/related"
	ctTextPrefix       = "text/"
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
	headerDeclarationRegex, err := regexp.Compile("^[a-zA-Z0-9\\-]+\\: ")
	if err != nil {
		buf.Write([]byte{'\r', '\n'})
		return nil, errors.WithStack(err)
	}
	for {
		// Pull out each line of the headers as a temporary slice s
		s, err := tp.ReadLineBytes()
		if err != nil {
			buf.Write([]byte{'\r', '\n'})
			break
		}
		firstColon := bytes.IndexByte(s, ':')
		firstSpace := bytes.IndexAny(s, " \t\n\r")
		if firstSpace == 0 {
			// If the line begins with some space, followed by a header-like
			// string (any combination of upper and lower case letters,
			// numbers and dash sign), then it should not be considered
			// as a continuation but as a new header.
			sTrimmed := textproto.TrimBytes(s)
			if firstSpace < firstColon && headerDeclarationRegex.Match(sTrimmed) {
				firstColon = bytes.IndexByte(s, ':')
			} else {
				// Starts with space: continuation
				buf.WriteByte(' ')
				buf.Write(sTrimmed)
				continue
			}
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

			// Behavior change in net/textproto package in Golang 1.12.10 and 1.13.1:
			// A space preceding the first colon in a header line is no longer handled
			// automatically due to CVE-2019-16276 which takes advantage of this
			// particular violation of RFC-7230 to exploit HTTP/1.1
			if bytes.Contains(s[:firstColon+1], []byte{' ', ':'}) {
				s = bytes.Replace(s, []byte{' ', ':'}, []byte{':'}, 1)
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

	// The standard lib performs an incremental inspection of this string, where the
	// "skipSpace" method only strings.trimLeft for spaces and tabs. Here we have a
	// hard dependency on space existing and not on next expected rune
	//
	// For resolving #112 with the least change, I will implement the
	// "quoted display-name" detector, which will resolve the case specific
	// issue stated in #112, but only in the case of a quoted display-name
	// followed, without whitespace, by addr-spec.
	tokens := strings.FieldsFunc(quotedDisplayName(input), whiteSpaceRune)
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

func quotedDisplayName(s string) string {
	if !strings.HasPrefix(s, "\"") {
		return s
	}
	idx := strings.LastIndex(s, "\"")
	return fmt.Sprintf("%s %s", s[:idx+1], s[idx+1:])
}

// ParseMediaType is a more tolerant implementation of Go's mime.ParseMediaType function.
//
// Tolerances accounted for:
//   * Missing ';' between content-type and media parameters
//   * Repeating media parameters
//   * Unquoted values in media parameters containing 'tspecials' characters
func ParseMediaType(ctype string) (mtype string, params map[string]string, invalidParams []string, err error) {
	mtype, params, err = mime.ParseMediaType(fixUnescapedQuotes(fixUnquotedSpecials(fixMangledMediaType(ctype, ";"))))
	if err != nil {
		if err.Error() == "mime: no media type" {
			return "", nil, nil, nil
		}
		return "", nil, nil, errors.WithStack(err)
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
	if strings.Contains(parts[0], "=") {
		// A parameter pair at this position indicates we are missing a content-type.
		parts[0] = fmt.Sprintf("%s%s %s", ctAppOctetStream, sep, parts[0])
		parts = strings.Split(strings.Join(parts, sep), sep)
	}
	for i, p := range parts {
		switch i {
		case 0:
			if p == "" {
				// The content type is completely missing. Put in a placeholder.
				p = ctPlaceholder
			}
			// Check for missing token after slash
			if strings.HasSuffix(p, "/") {
				switch p {
				case ctTextPrefix:
					p = ctTextPlain
				case ctAppPrefix:
					p = ctAppOctetStream
				case ctMultipartPrefix:
					p = ctMultipartMixed
				default:
					// Safe default
					p = ctAppOctetStream
				}
			}
		default:
			if len(p) == 0 {
				// Ignore trailing separators.
				continue
			}

			if !strings.Contains(p, "=") {
				p = p + "=" + pvPlaceholder
			}

			// RFC-2047 encoded attribute name
			p = rfc2047decode(p)

			pair := strings.SplitAfter(p, "=")
			if strings.Contains(mtype, strings.TrimSpace(pair[0])) {
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

		case ';':
			if value.Len() == 0 {
				value.WriteString(`"";`)
			}

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

		quoteIfUnquoted := func() {
			if !valueQuoteNeeded {
				if !valueQuoteAdded {
					param.WriteByte('"')

					valueQuoteAdded = true
				}

				valueQuoteNeeded = true
			}
		}

		for _, v := range []byte{'(', ')', '<', '>', '@', ',', ':', '/', '[', ']', '?', '='} {
			if s[0] == v {
				quoteIfUnquoted()
			}
		}

		s = s[i+1:]

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

// fixUnescapedQuotes inspects for unescaped quotes inside of a quoted string and escapes them
//
//  Input:  application/rtf; charset=iso-8859-1; name=""V047411.rtf".rtf"
//  Output: application/rtf; charset=iso-8859-1; name="\"V047411.rtf\".rtf"
func fixUnescapedQuotes(hvalue string) string {
	params := strings.SplitAfter(hvalue, ";")
	sb := &strings.Builder{}
	for i := 0; i < len(params); i++ {
		// Inspect for "=" byte.
		eq := strings.IndexByte(params[i], '=')
		if eq < 0 {
			// No "=", must be the content-type or a comment.
			sb.WriteString(params[i])
			continue
		}
		sb.WriteString(params[i][:eq])
		param := params[i][eq:]
		startingQuote := strings.IndexByte(param, '"')
		closingQuote := strings.LastIndexByte(param, '"')
		// Opportunity to exit early if there are no quotes.
		if startingQuote < 0 && closingQuote < 0 {
			// This value is not quoted, write the value and carry on.
			sb.WriteString(param)
			continue
		}
		// Check if only one quote was found in the string.
		if closingQuote == startingQuote {
			// Append the next chunk of params here in case of a semicolon mid string.
			if len(params) > i+1 {
				param = fmt.Sprintf("%s%s", param, params[i+1])
			}
			closingQuote = strings.LastIndexByte(param, '"')
			i++
			if closingQuote == startingQuote {
				sb.WriteString("=\"\"")
				return sb.String()
			}
		}
		// Write the k/v separator back in along with everything up until the first quote.
		sb.WriteByte('=')
		// Starting quote
		sb.WriteByte('"')
		sb.WriteString(param[1:startingQuote])
		// Get just the value, less the outer quotes.
		rest := param[closingQuote+1:]
		// If there is stuff after the last quote then we should escape the first quote.
		if len(rest) > 0 && rest != ";" {
			sb.WriteString("\\\"")
		}
		param = param[startingQuote+1 : closingQuote]
		escaped := false
		for strIdx := range []byte(param) {
			switch param[strIdx] {
			case '"':
				// We are inside of a quoted string, so lets escape this guy if it isn't already escaped.
				if !escaped {
					sb.WriteByte('\\')
					escaped = false
				}
				sb.WriteByte(param[strIdx])
			case '\\':
				// Something is getting escaped, a quote is the only char that needs
				// this, so lets assume the following char is a double-quote.
				escaped = true
				sb.WriteByte('\\')
			default:
				escaped = false
				sb.WriteByte(param[strIdx])
			}
		}
		// If there is stuff after the last quote then we should escape
		// the last quote, apply the rest and terminate with a quote.
		switch rest {
		case ";":
			sb.WriteByte('"')
			sb.WriteString(rest)
		case "":
			sb.WriteByte('"')
		default:
			sb.WriteByte('\\')
			sb.WriteByte('"')
			sb.WriteString(rest)
			sb.WriteByte('"')
		}
	}
	return sb.String()
}

// Detects a RFC-822 linear-white-space, passed to strings.FieldsFunc.
func whiteSpaceRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}
