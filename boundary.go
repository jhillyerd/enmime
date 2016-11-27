package enmime

import (
	"bufio"
	"bytes"
	"io"
)

// This constant needs to be at least 76 for this package to work correctly.  This is because
// \r\n--separator_of_len_70- would fill the buffer and it wouldn't be safe to consume a single byte
// from it.
// TODO implement test with boundary crossing edge of buffer
const peekBufferSize = 4096

type boundaryReader struct {
	r      *bufio.Reader // Source reader
	prefix []byte        // Boundary prefix we are looking for
	buffer *bytes.Buffer
}

func newBoundaryReader(reader *bufio.Reader, boundary string) *boundaryReader {
	return &boundaryReader{
		r:      reader,
		prefix: []byte("\n--" + boundary),
		buffer: new(bytes.Buffer),
	}
}

// readUntilBoundary returns a buffer containing the content up until boundary
func (b *boundaryReader) Read(dest []byte) (n int, err error) {
	if b.buffer.Len() >= len(dest) {
		// We can satify this request with buffer only
		return b.buffer.Read(dest)
	}

	peek, err := b.r.Peek(peekBufferSize)
	peekEOF := (err == io.EOF)
	if err != nil && !peekEOF && err != bufio.ErrBufferFull {
		return 0, err
	}
	var nCopy int
	idx, complete := locateBoundary(peek, b.prefix)
	if idx != -1 {
		// Peeked boundary prefix, read until that point
		nCopy = idx
		if !complete && nCopy == 0 {
			// Incomplete boundary, move past it
			nCopy = 1
		}
	} else {
		// No boundary found, read peek minus the length of boundaryPrefix, minus one more for
		// potential \r
		if nCopy = len(peek) - len(b.prefix) - 1; nCopy <= 0 {
			nCopy = 0
			if peekEOF {
				// We've run out of peek space with no boundary found
				return 0, io.ErrUnexpectedEOF
			}
		}
	}
	if nCopy > 0 {
		if _, err := io.CopyN(b.buffer, b.r, int64(nCopy)); err != nil {
			return 0, err
		}
	}

	n, err = b.buffer.Read(dest)
	if err == io.EOF && !complete {
		// buffer is empty, but not the boundaryReader
		return n, nil
	}
	return
}

// Locate boundaryPrefix in buf, returning its starting idx. If complete is true, the boundary
// is terminated properly in buf, otherwise it could be false due to running out of buffer, or
// because it is not the actual boundary.
//
// Complete boundaries end in "--" or a newline
func locateBoundary(buf, boundaryPrefix []byte) (idx int, complete bool) {
	bpLen := len(boundaryPrefix)
	idx = bytes.Index(buf, boundaryPrefix)
	if idx == -1 {
		return
	}

	// Handle CR if present
	if idx > 0 && buf[idx-1] == '\r' {
		idx--
		bpLen++
	}

	// Fast forward to the end of the boundary prefix
	buf = buf[idx+bpLen:]
	if len(buf) == 0 {
		// Need more bytes to verify completeness
		return
	}
	if len(buf) > 0 {
		if buf[0] == '\r' || buf[0] == '\n' {
			return idx, true
		}
	}
	if len(buf) > 1 {
		if buf[0] == '-' && buf[1] == '-' {
			return idx, true
		}
	}

	return idx, false
}
