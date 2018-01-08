package enmime

import (
	"fmt"
	"io"
)

// base64CleanerTable notes byte values that should be stripped (-2), stripped w/ error (-1)
var base64CleanerTable = []int8{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -2, -2, -1, -1, -2, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-2, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 62, -1, -1, -1, 63,
	52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, -2, -1, -1,
	-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
	15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -1, -1, -1, -1, -1,
	-1, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
	41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, -1, -1, -1, -1, -1,
}

// base64Cleaner helps work around bugs in Go's built-in base64 decoder by stripping out
// whitespace that would cause Go to lose count of things and issue an "illegal base64 data at
// input byte..." error
type base64Cleaner struct {
	// Errors detected while cleaning base64 data
	Errors []Error

	r      io.Reader
	buffer [1024]byte
}

// Enforce io.Reader interface
var _ io.Reader = &base64Cleaner{}

// newBase64Cleaner returns a Base64Cleaner object for the specified reader.  Base64Cleaner
// implements the io.Reader interface.
func newBase64Cleaner(r io.Reader) *base64Cleaner {
	return &base64Cleaner{
		Errors: make([]Error, 0),
		r:      r,
	}
}

// Read method for io.Reader interface.
func (bc *base64Cleaner) Read(p []byte) (n int, err error) {
	// Size our buf to smallest of len(p) or len(bc.buffer)
	size := len(bc.buffer)
	if size > len(p) {
		size = len(p)
	}
	buf := bc.buffer[:size]
	bn, err := bc.r.Read(buf)
	for i := 0; i < bn; i++ {
		switch base64CleanerTable[buf[i]&0x7f] {
		case -2:
			// Strip these silently (tab, \n, \r, space, =)
		case -1:
			// Strip these, but warn the client
			bc.Errors = append(bc.Errors, Error{
				Name:   ErrorMalformedBase64,
				Detail: fmt.Sprintf("Unexpected %q in Base64 stream", buf[i]),
				Severe: false,
			})
		default:
			p[n] = buf[i]
			n++
		}
	}
	return
}
