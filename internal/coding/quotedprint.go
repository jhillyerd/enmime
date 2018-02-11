package coding

import (
	"bufio"
	"fmt"
	"io"
)

// QPCleaner scans quoted printable content for invalid characters and encodes them so that
// Go's quoted-printable decoder does not abort with an error.
type QPCleaner struct {
	in *bufio.Reader
}

// Assert QPCleaner implements io.Reader.
var _ io.Reader = &QPCleaner{}

// NewQPCleaner returns a QPCleaner for the specified reader.
func NewQPCleaner(r io.Reader) *QPCleaner {
	return &QPCleaner{
		in: bufio.NewReader(r),
	}
}

// Read method for io.Reader interface.
func (qp *QPCleaner) Read(dest []byte) (n int, err error) {
	// Ensure room to write a byte or =XX string.
	destLen := len(dest) - 3
	// Loop over bytes in qp.in ByteReader.
	for n < destLen {
		b, err := qp.in.ReadByte()
		if err != nil {
			return n, err
		}
		switch {
		case b == '=':
			// Pass valid hex bytes through.
			hexBytes, err := qp.in.Peek(2)
			if err != nil && err != io.EOF {
				return 0, err
			}
			if validHexBytes(hexBytes) {
				dest[n] = b
				n++
			} else {
				s := fmt.Sprintf("=%02X", b)
				n += copy(dest[n:], s)
			}
		case b == '\t' || b == '\r' || b == '\n':
			// Valid special character.
			dest[n] = b
			n++
		case b < ' ' || '~' < b:
			// Invalid character, render quoted-printable into buffer.
			s := fmt.Sprintf("=%02X", b)
			n += copy(dest[n:], s)
		default:
			// Acceptable character.
			dest[n] = b
			n++
		}
	}
	return n, err
}

func validHexByte(b byte) bool {
	return '0' <= b && b <= '9' || 'A' <= b && b <= 'F' || 'a' <= b && b <= 'f'
}

func validHexBytes(v []byte) bool {
	if len(v) < 1 {
		return false
	}
	if v[0] == '\n' {
		// Soft line break.
		return true
	}
	if len(v) < 2 {
		return false
	}
	if v[0] == '\r' && v[1] == '\n' {
		// Soft line break.
		return true
	}
	return validHexByte(v[0]) && validHexByte(v[1])
}
