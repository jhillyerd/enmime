package enmime

import (
	"bufio"
	"bytes"
	"fmt"
	"mime"
	"net/textproto"
	"strings"

	"github.com/jhillyerd/enmime/internal/coding"
	"github.com/jhillyerd/enmime/mediatype"
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

// ParseMediaType is a more tolerant implementation of Go's mime.ParseMediaType function.
//
// Tolerances accounted for:
//   * Missing ';' between content-type and media parameters
//   * Repeating media parameters
//   * Unquoted values in media parameters containing 'tspecials' characters
func ParseMediaType(ctype string) (mtype string, params map[string]string, invalidParams []string,
	err error) {
	// Export of internal function.
	return mediatype.Parse(ctype)
}

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
			buf.Write([]byte{'\r', '\n'})
			break
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

// decodeToUTF8Base64Header decodes a MIME header per RFC 2047, reencoding to =?utf-8b?
func decodeToUTF8Base64Header(input string) string {
	if !strings.Contains(input, "=?") {
		// Don't scan if there is nothing to do here
		return input
	}

	// The standard lib performs an incremental inspection of this string, where the
	// "skipSpace" method only strings.trimLeft for spaces and tabs. Here we have a
	// hard dependency on space existing and not on next expected rune.
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
			output[i] = prefix +
				mime.BEncoding.Encode("UTF-8", coding.DecodeExtHeader(token)) +
				suffix
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

// Detects a RFC-822 linear-white-space, passed to strings.FieldsFunc.
func whiteSpaceRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}
