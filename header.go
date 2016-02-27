package enmime

import (
	"fmt"
	"mime"
	"strings"
)

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

// DecodeHeader (per RFC 2047) using Golang's mime.WordDecoder
func DecodeHeader(input string) string {
	if !strings.Contains(input, "=?") {
		// Don't scan if there is nothing to do here
		return input
	}

	dec := new(mime.WordDecoder)
	dec.CharsetReader = NewCharsetReader
	header, err := dec.DecodeHeader(input)
	if err != nil {
		return input
	}
	return header
}

// DecodeToUTF8Base64Header decodes a MIME header per RFC 2047, reencoding to =?utf-8b?
func DecodeToUTF8Base64Header(input string) string {
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
			output[i] = prefix + mime.BEncoding.Encode("UTF-8", DecodeHeader(token)) + suffix
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
