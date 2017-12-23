package enmime

import (
	"io"
)

// something inspired by table_a2b_base64 in binascii module of python
var base64CleanerTable = []int8{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 62, -1, -1, -1, 63,
	52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, -1, -1, -1,
	-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
	15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -1, -1, -1, -1, -1,
	-1, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
	41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, -1, -1, -1, -1, -1,
}

// base64Cleaner helps work around bugs in Go's built-in base64 decoder by stripping out
// whitespace that would cause Go to lose count of things and issue an "illegal base64 data at
// input byte..." error
type base64Cleaner struct {
	in  io.Reader
	buf [1024]byte
	//count int64
}

// newBase64Cleaner returns a Base64Cleaner object for the specified reader.  Base64Cleaner
// implements the io.Reader interface.
func newBase64Cleaner(r io.Reader) *base64Cleaner {
	return &base64Cleaner{in: r}
}

// Read method for io.Reader interface.
func (qp *base64Cleaner) Read(p []byte) (n int, err error) {
	// Size our slice to theirs
	size := len(qp.buf)
	if len(p) < size {
		size = len(p)
	}
	buf := qp.buf[:size]
	bn, err := qp.in.Read(buf)
	for i := 0; i < bn; i++ {
		switch base64CleanTable[buf[i]&0x7f] {
		case -1:
			// Strip these
		default:
			p[n] = buf[i]
			n++
		}
	}
	// Count may be useful if I need to pad to even quads
	//qp.count += int64(n)
	return n, err
}
