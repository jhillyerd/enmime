package enmime

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Terminology from RFC 2047:
//  encoded-word: the entire =?charset?encoding?encoded-text?= string
//  charset: the character set portion of the encoded word
//  encoding: the character encoding type used for the encoded-text
//  encoded-text: the text we are decoding
const (
	PLAIN    int = iota // In plain text
	RESET               // Parse error, output encoded-word as plaintext
	CHARSET             // In charset name
	ENCODING            // In encoding name
	ENCTEXT             // In encoded-text
	SPACE               // Space following an encoded-word
)

// Decode a MIME header per RFC 2047
func decodeHeader(input string) (utf8 string, err error) {
	if ! strings.Contains(input, "=?") {
		// Don't scan if there is nothing to do here
		return input, nil
	}

	outbuf := new(bytes.Buffer)
	state := PLAIN
	startPos := 0
	var charsetBytes, encodingBytes, encTextBytes []byte

	for pos := 0; pos < len(input); pos++ {
		ch := input[pos]
		switch state {
		case SPACE:
			// Eat space characters only between encoded words
			switch ch {
			case ' ', '\t', '\r', '\n':
				// Still in space
				continue
			case '=':
				if len(input) > pos+1 && input[pos+1] == '?' {
					// Start of new encoded word.  If the word is valid, we want to eat
					// the whitespace.  If not, startPos was defined in transition to SPACE,
					// and we will output it.
					pos++
					state = CHARSET
					charsetBytes = make([]byte, 0, 128)
					continue
				}
			}
			// We hit plain text, will need to output whitespace
			for i := startPos; i <= pos; i++ {
				outbuf.WriteByte(input[i])
			}
			state = PLAIN
		case PLAIN:
			// Scan for start of encoded word: =?
			if ch == '=' && len(input) > pos+1 && input[pos+1] == '?' {
				// Save start in case this turns out to be corrupt
				startPos = pos
				pos++
				state = CHARSET
				charsetBytes = make([]byte, 0, 128)
				continue
			}

			// Plain text
			outbuf.WriteByte(ch)
		case RESET:
			// There was a problem parsing an encoded-word, recover by outputting
			// plain-text until next encoded word
			for i := startPos; i <= pos; i++ {
				outbuf.WriteByte(input[i])
			}
			state = PLAIN
		case CHARSET:
			// Parse character set
			switch {
			case isTokenChar(ch):
				// Part of charset name
				charsetBytes = append(charsetBytes, ch)
			case ch == '?':
				// End of charset name
				state = ENCODING
				encodingBytes = make([]byte, 0, 128)
			default:
				state = RESET
			}
		case ENCODING:
			// Parse encoding
			switch {
			case isTokenChar(ch):
				// Part of encoding name
				encodingBytes = append(encodingBytes, ch)
			case ch == '?':
				// End of encoding
				state = ENCTEXT
				encTextBytes = make([]byte, 0, 128)
			default:
				state = RESET
			}
		case ENCTEXT:
			// Decode encoded-text
			switch {
			case ch < 33:
				// No controls or space allowed
				state = RESET
			case ch > 126:
				// No DEL or extended ascii allowed
				state = RESET
			case ch == '?':
				if len(input) > pos+1 && input[pos+1] == '=' {
					text, err := convertText(charsetBytes, encodingBytes, encTextBytes)
					if err == nil {
						outbuf.WriteString(text)
						pos++
						// Entering post-word space
						state = SPACE
						startPos = pos+1
					} else {
						// Conversion failed
						state = RESET
					}
				} else {
					// Invalid termination
					state = RESET
				}
			default:
				encTextBytes = append(encTextBytes, ch)
			}
		}
	}

	return outbuf.String(), nil
}

// Convert the encTextBytes to UTF-8 and return as a string
func convertText(charsetBytes, encodingBytes, encTextBytes []byte) (string, error) {
	encoding := strings.ToLower(string(encodingBytes))
	switch encoding {
	case "b":
		// Base64 encoded
		textBytes, err := decodeBase64(encTextBytes)
		return string(textBytes), err
	case "q":
		// Quoted printable encoded
		textBytes, err := decodeQuotedPrintable(encTextBytes)
		return string(textBytes), err
	default:
		return "", fmt.Errorf("Invalid encoding: %v", encoding)
	}	
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
func isEspecialChar(ch byte) bool {
	switch ch {
	case '(', ')', '<', '>', '@', ',', ';', ':':
		return true
	case '"', '/', '[', ']', '?', '.', '=':
		return true
	}
	return false
}

// Is this a "token" character per RFC 2047
func isTokenChar(ch byte) bool {
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
