package enmime

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
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

// State function modeled on Rob Pike's lexer talk (see source of
// Go's text/template/parser/lex.go)
type stateFn func(*headerDec) stateFn

const eof = -1

// headerDec holds the state of the scanner and an output buffer
type headerDec struct {
	input    []byte  // Input to decode
	state    stateFn // Next state
	start    int     // Start of text we don't yet know what to do with
	pos      int     // Current parsing position
	charset  string  // Character set of current encoded word
	encoding string  // Encoding of current encoded word
	trans    bool    // Convert other character to utf8-base64 =?UTF-8?B?==?=
	outbuf   bytes.Buffer
}

// eof returns true if we've read the last rune
func (h *headerDec) eof() bool {
	return h.pos >= len(h.input)
}

// next returns the next rune in the input
func (h *headerDec) next() rune {
	if h.eof() {
		return eof
	}
	r := h.input[h.pos]
	h.pos++
	return rune(r)
}

// backup a single rune
func (h *headerDec) backup() {
	if h.pos > 0 {
		h.pos--
	}
}

// peek at the next rune without consuming it
func (h *headerDec) peek() rune {
	if h.eof() {
		return eof
	}
	r := h.next()
	h.backup()
	return r
}

// ignore will forget all input between start and pos
func (h *headerDec) ignore() {
	h.start = h.pos
}

// output will append all input from start to pos (inclusive) to outbuf
func (h *headerDec) output() {
	if h.pos > h.start {
		h.outbuf.Write(h.input[h.start:h.pos])
		h.start = h.pos
	}
}

// accept consumes the next rune if it's part of the valid set
func (h *headerDec) accept(valid string) bool {
	if r := h.next(); r != eof {
		if strings.IndexRune(valid, r) >= 0 {
			return true
		}
		h.backup()
	}
	return false
}

// Decode a MIME header per RFC 2047
func DecodeHeader(input string) string {
	if !strings.Contains(input, "=?") {
		// Don't scan if there is nothing to do here
		return input
	}

	h := &headerDec{
		input: []byte(input),
		state: plainSpaceState,
	}

	debug("Starting parse of: '%v'\n", input)

	for h.state != nil {
		h.state = h.state(h)
	}

	return h.outbuf.String()
}

// Decode a MIME header per RFC 2047 to =?utf-8b?
func DecodeToUTF8Base64Header(input string) string {
	if !strings.Contains(input, "=?") {
		// Don't scan if there is nothing to do here
		return input
	}

	h := &headerDec{
		input: []byte(input),
		state: plainSpaceState,
		trans: true,
	}

	debug("Starting parse of: '%v'\n", input)

	for h.state != nil {
		h.state = h.state(h)
	}

	return h.outbuf.String()
}

// State: Reset, output mangled encoded-word as plaintext
//
// There was a problem parsing an encoded-word, recover by outputting
// plain-text until next encoded word
func resetState(h *headerDec) stateFn {
	debug("entering reset state with buf %q", h.outbuf.String())
	h.output()
	return plainTextState
}

// State: In plain space - we want to output this space, and it is legal to transition into
// an encoded word from here.
func plainSpaceState(h *headerDec) stateFn {
	debug("entering plain space state with buf %q", h.outbuf.String())
	for !h.eof() {
		switch {
		case h.accept("="):
			// Possible encoded word, dump out preceeding plaintext, w/o leading =
			h.backup()
			h.output()
			h.next()
			if h.accept("?") {
				return charsetState
			}
		case h.accept(" \t\r\n("):
			// '(' is the start of a MIME header "comment", so it is still legal to transition
			// to an encoded word after it
		default:
			// Hit some plain text
			return plainTextState
		}
	}
	// Hitting EOF in plain space state means we are done
	h.output()
	return nil
}

// State: In plain text - we want to output this text.  It is not legal to transition into
// an encoded word from here!
func plainTextState(h *headerDec) stateFn {
	debug("entering plain text state with buf %q", h.outbuf.String())
	for !h.eof() {
		if h.accept(" \t\r\n(") {
			// TODO Not sure if '(' belongs here, maybe require space first?
			// Whitespace character
			h.backup()
			return plainSpaceState
		}
		h.next()
	}
	// Hitting EOF in plain text state means we are done
	h.output()
	return nil
}

// State: In encoded-word charset name
func charsetState(h *headerDec) stateFn {
	debug("entering charset state with buf %q", h.outbuf.String())
	myStart := h.pos
	for r := h.next(); r != eof; r = h.next() {
		// Parse character set
		switch {
		case isTokenChar(r):
			// Part of charset name, keep going
		case r == '?':
			// End of charset name
			h.charset = string(h.input[myStart : h.pos-1])
			debug("charset %q", h.charset)
			return encodingState
		default:
			// Invalid character
			return resetState
		}
	}
	// Hit eof!
	return resetState
}

// State: In encoded-word encoding name
func encodingState(h *headerDec) stateFn {
	debug("entering encoding state with buf %q", h.outbuf.String())
	myStart := h.pos
	for r := h.next(); r != eof; r = h.next() {
		// Parse encoding
		switch {
		case isTokenChar(r):
			// Part of encoding name, keep going
		case r == '?':
			// End of encoding name
			h.encoding = string(h.input[myStart : h.pos-1])
			debug("encoding %q", h.encoding)
			return encTextState
		default:
			// Invalid character
			debug("invalid character")
			return resetState
		}
	}
	// Hit eof!
	debug("hit unexpected eof")
	return resetState
}

// State: In encoded-text
func encTextState(h *headerDec) stateFn {
	debug("entering encText state with buf %q", h.outbuf.String())
	myStart := h.pos
	for r := h.next(); r != eof; r = h.next() {
		// Decode encoded-text
		switch {
		case r < 33:
			// No controls or space allowed
			debug("Encountered control character")
			return resetState
		case r > 126:
			// No DEL or extended ascii allowed
			debug("Encountered DEL or extended ascii")
			return resetState
		case r == '?':
			if h.accept("=") {
				text, err := convertText(h.charset, h.encoding, h.input[myStart:h.pos-2])
				if err == nil {
					debug("Text converted to: %q", text)
					if h.trans {
						h.outbuf.WriteString("=?UTF-8?B?")
						h.outbuf.WriteString(base64.StdEncoding.EncodeToString(([]byte)(text)))
						h.outbuf.WriteString("?=")
					} else {
						h.outbuf.WriteString(text)
					}
					h.ignore()
					// Entering post-word space
					return spaceState
				} else {
					// Conversion failed
					debug("Text conversion failed: %q", err)
					return resetState
				}
			} else {
				// Invalid termination
				debug("Invalid termination")
				return resetState
			}
		}
	}
	// Hit eof!
	debug("Hit eof early")
	return resetState
}

// State: White space following an encoded-word
func spaceState(h *headerDec) stateFn {
	debug("entering space state with buf %q", h.outbuf.String())
Loop:
	for {
		// Eat space characters only between encoded words
		switch {
		case h.accept(" \t\r\n"):
			debug("In space")
			// Still in space
		case h.accept("="):
			debug("In =")
			if h.accept("?") {
				// Start of new encoded word.  If the word is valid, we want to eat
				// the whitespace.  If not, h.start was set in transition to SPACE,
				// and we will output the space.
				return charsetState
			}
		default:
			debug("In default")
			break Loop
		}
	}
	debug("In plain")
	// We hit plain text, will need to output whitespace
	h.output()
	return plainTextState
}

// Convert the encTextBytes to UTF-8 and return as a string
func convertText(charset string, encoding string, encTextBytes []byte) (string, error) {
	// Unpack quoted-printable or base64 first
	var err error
	var textBytes []byte
	switch strings.ToLower(encoding) {
	case "b":
		// Base64 encoded
		textBytes, err = decodeBase64(encTextBytes)
	case "q":
		// Quoted printable encoded
		textBytes, err = decodeQuotedPrintable(encTextBytes)
	default:
		err = fmt.Errorf("Invalid encoding: %v", encoding)
	}
	if err != nil {
		return "", err
	}

	return ConvertToUTF8String(charset, string(textBytes))
}

func decodeQuotedPrintable(input []byte) ([]byte, error) {
	output := make([]byte, 0, len(input))
	for pos := 0; pos < len(input); pos++ {
		switch ch := input[pos]; ch {
		case '_':
			output = append(output, ' ')
		case '=':
			if len(input) < pos+3 {
				return nil, fmt.Errorf("Ran out of chars parsing: %v", input[pos:])
			}
			x, err := strconv.ParseInt(string(input[pos+1:pos+3]), 16, 64)
			if err != nil {
				return nil, fmt.Errorf("Failed to convert: %v", input[pos:pos+3])
			}
			output = append(output, byte(x))
			pos += 2
		default:
			output = append(output, input[pos])
		}
	}
	return output, nil
}

func decodeBase64(input []byte) ([]byte, error) {
	output := make([]byte, len(input))
	n, err := base64.StdEncoding.Decode(output, input)
	return output[:n], err
}

// Is this an especial character per RFC 2047
func isEspecialChar(ch rune) bool {
	switch ch {
	case '(', ')', '<', '>', '@', ',', ';', ':':
		return true
	case '"', '/', '[', ']', '?', '.', '=':
		return true
	}
	return false
}

// Is this a "token" character per RFC 2047
func isTokenChar(ch rune) bool {
	// No controls or space
	if ch < 33 {
		return false
	}
	// No DEL or extended ascii
	if ch > 126 {
		return false
	}
	// No especials
	return !isEspecialChar(ch)
}
