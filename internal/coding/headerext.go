package coding

import (
	"mime"
	"strings"
)

// DecodeExtHeader decodes a single line (per RFC 2047, aka Message Header Extensions) using Golang's
// mime.WordDecoder
func DecodeExtHeader(input string) string {
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
