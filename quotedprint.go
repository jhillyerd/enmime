package enmime

import (
	"bufio"
	"fmt"
	"io"
)

// qpCleaner scans quoted printable content for invalid characters and encodes them so that
// Go's quoted-printable decoder does not abort with an error.
type qpCleaner struct {
	in *bufio.Reader
}

// Assert qpCleaner implements io.Reader
var _ io.Reader = &qpCleaner{}

// newBase64Cleaner returns a Base64Cleaner object for the specified reader.  Base64Cleaner
// implements the io.Reader interface.
func newQPCleaner(r io.Reader) *qpCleaner {
	return &qpCleaner{
		in: bufio.NewReader(r),
	}
}

// Read method for io.Reader interface.
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
		case b == '=':
			// pass valid hex bytes through
			hexBytes, err := qp.in.Peek(2)
			if err != nil && err != io.EOF {
				return 0, err
			}
			if isValidHexBytes(hexBytes) {
				dest[n] = b
				n++
			} else {
				s := fmt.Sprintf("=%02X", b)
				n += copy(dest[n:], s)
			}
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

func isValidHexByte(b byte) bool {
	switch {
	case b >= '0' && b <= '9':
		return true
	case b >= 'A' && b <= 'F':
		return true
	// Accept badly encoded bytes.
	case b >= 'a' && b <= 'f':
		return true
	}
	return false
}

func isValidHexBytes(v []byte) bool {
	if len(v) < 1 {
		return false
	}

	// soft line break
	if v[0] == '\n' {
		return true
	}

	if len(v) < 2 {
		return false
	}

	// soft line break
	if v[0] == '\r' && v[1] == '\n' {
		return true
	}

	if isValidHexByte(v[0]) && isValidHexByte(v[1]) {
		return true
	}

	return false
}
