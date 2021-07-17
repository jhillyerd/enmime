package coding

import (
	"bufio"
	"fmt"
	"io"
)

// QPCleaner scans quoted printable content for invalid characters and encodes them so that
// Go's quoted-printable decoder does not abort with an error.
type QPCleaner struct {
	in       *bufio.Reader
	overflow []byte
}

// Assert QPCleaner implements io.Reader.
var _ io.Reader = &QPCleaner{}

// escapedEquals is the QP encoded value of an equals sign.
var escapedEquals = "=3D"

// NewQPCleaner returns a QPCleaner for the specified reader.
func NewQPCleaner(r io.Reader) *QPCleaner {
	return &QPCleaner{
		in:       bufio.NewReader(r),
		overflow: nil,
	}
}

// Read method for io.Reader interface.
func (qp *QPCleaner) Read(dest []byte) (n int, err error) {
	destLen := len(dest)

	if len(qp.overflow) > 0 {
		// Copy bytes that didn't fit into dest buffer during previous read.
		n = copy(dest, qp.overflow)
		qp.overflow = qp.overflow[n:]
	}

	// Loop over bytes in qp.in ByteReader.
	for n < destLen {
		var b byte
		b, err = qp.in.ReadByte()
		if err != nil {
			return n, err
		}

		switch {
		case b == '=':
			// Pass valid hex bytes through.
			var hexBytes []byte
			hexBytes, err = qp.in.Peek(2)
			if err != nil && err != io.EOF {
				return 0, err
			}
			if validHexBytes(hexBytes) {
				dest[n] = b
				n++
			} else {
				nc := copy(dest[n:], escapedEquals)
				if nc < len(escapedEquals) {
					// Stash unwritten bytes into overflow.
					qp.overflow = []byte(escapedEquals[nc:])
				}
				n += nc
			}

		case b == '\t' || b == '\r' || b == '\n':
			// Valid special character.
			dest[n] = b
			n++

		case b < ' ' || '~' < b:
			// Invalid character, render quoted-printable into buffer.
			s := fmt.Sprintf("=%02X", b)
			nc := copy(dest[n:], s)
			if nc < len(s) {
				// Stash unwritten bytes into overflow.
				qp.overflow = []byte(s[nc:])
			}
			n += nc

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

// validHexBytes returns true if this byte sequence represents a valid quoted-printable escape
// sequence or line break, minus the initial equals sign.
func validHexBytes(v []byte) bool {
	if len(v) > 0 && v[0] == '\n' {
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
