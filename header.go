package enmime

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/textproto"
	"strings"
)

const (
	// Standard MIME content dispositions
	cdAttachment = "attachment"
	cdInline     = "inline"

	// Standard MIME content types
	ctAppOctetStream  = "application/octet-stream"
	ctMultipartAltern = "multipart/alternative"
	ctMultipartPrefix = "multipart/"
	ctTextPlain       = "text/plain"
	ctTextHTML        = "text/html"

	// Standard MIME header names
	hnContentDisposition = "Content-Disposition"
	hnContentEncoding    = "Content-Transfer-Encoding"
	hnContentType        = "Content-Type"

	// Standard MIME header parameters
	hpBoundary = "boundary"
	hpCharset  = "charset"
	hpFile     = "file"
	hpFilename = "filename"
	hpName     = "name"
)

var errEmptyHeaderBlock = errors.New("empty header block")

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

func debug(format string, args ...interface{}) {
	if false {
		fmt.Printf(format, args...)
		fmt.Println()
	}
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

	continuable := false
	for {
		// Pull out each line of the headers as a temporary slice s
		s, err := r.ReadSlice('\n')
		if err != nil {
			if err == io.ErrUnexpectedEOF && buf.Len() == 0 {
				return nil, errEmptyHeaderBlock
			}
			return nil, err
		}
		// Remove trailing whitespace
		s = bytes.TrimRight(s, " \t\r\n")
		if len(s) == 0 {
			// End of headers
			break
		}
		if continuable {
			// Attempt to detect and repair a non-indented continuation of previous line
			if s[0] != ' ' && s[0] != '\t' {
				// Not already indented; if the first word does not end in a colon, this is a
				// mangled continuation
				firstSpace := bytes.IndexByte(s, ' ')
				firstColon := bytes.IndexByte(s, ':')
				if (firstColon < 0) || (0 <= firstSpace && firstSpace < firstColon) {
					// Continuation scenario, prepend line with space
					// "word word" (no colon)
					// "word word:word" (colon after space)
					buf.WriteByte(' ')
					p.addWarning(errorMalformedHeader, "Continued line %q was not indented", s)
				}
			}
			continuable = false
		}
		if s[len(s)-1] == ';' {
			// Next line may be a continuation of this one
			continuable = true
		}
		buf.Write(s)
		buf.Write([]byte{'\r', '\n'})
	}
	buf.Write([]byte{'\r', '\n'})

	// Parse the buf using textproto package
	tr := textproto.NewReader(bufio.NewReader(buf))
	header, err := tr.ReadMIMEHeader()
	return header, err
}

// decodeHeader decodes a single line (per RFC 2047) using Golang's mime.WordDecoder
func decodeHeader(input string) string {
	if !strings.Contains(input, "=?") {
		// Don't scan if there is nothing to do here
		return input
	}

	dec := new(mime.WordDecoder)
	dec.CharsetReader = newCharsetReader
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

	debug("input = %q", input)
	tokens := strings.FieldsFunc(input, isWhiteSpaceRune)
	output := make([]string, len(tokens), len(tokens))
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
		debug("%v %q %q", i, token, output[i])
	}

	// Return space separated tokens
	return strings.Join(output, " ")
}

// Detects a RFC-822 linear-white-space, passed to strings.FieldsFunc
func isWhiteSpaceRune(r rune) bool {
	switch r {
	case ' ':
		return true
	case '\t':
		return true
	case '\r':
		return true
	case '\n':
		return true
	default:
		return false
	}
}
