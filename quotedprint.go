package enmime

import (
	"fmt"
	"io"
)

// qpCleaner scans quoted printable content for invalid characters and encodes them so that
// Go's quoted-printable decoder does not abort with an error.
type qpCleaner struct {
	in io.ByteReader
}

// Assert qpCleaner implements io.Reader
var _ io.Reader = &qpCleaner{}

// newBase64Cleaner returns a Base64Cleaner object for the specified reader.  Base64Cleaner
// implements the io.Reader interface.
func newQPCleaner(r io.ByteReader) *qpCleaner {
	return &qpCleaner{
		in: r,
	}
}

// Read method for io.Reader interface.  Reasonably efficient for well-formed quoted-printable
// streams.  Less so when invalid characters are encountered; reads will be short.
func (qp *qpCleaner) Read(dest []byte) (n int, err error) {
	// Ensure room to write a byte or =XX string
	destLen := len(dest) - 3
	// Loop over bytes in qp.in ByteReader
	for n < destLen {
		b, err := qp.in.ReadByte()
		if err != nil {
			return n, err
		}
		// Test character type
		switch {
		case b == '\t' || b == '\r' || b == '\n':
			// Valid special characters
			dest[n] = b
			n++
		case b < ' ' || '~' < b:
			// Invalid character, render quoted-printable into buffer
			s := fmt.Sprintf("=%02X", b)
			n += copy(dest[n:], s)
		default:
			// Acceptable character
			dest[n] = b
			n++
		}
	}
	return
}
