package enmime

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

// This constant needs to be at least 76 for this package to work correctly.  This is because
// \r\n--separator_of_len_70- would fill the buffer and it wouldn't be safe to consume a single byte
// from it.
const peekBufferSize = 4096

type boundaryReader struct {
	finished  bool          // No parts remain when finished
	partsRead int           // Number of parts read thus far
	r         *bufio.Reader // Source reader
	nlPrefix  []byte        // NL + MIME boundary prefix
	prefix    []byte        // MIME boundary prefix
	final     []byte        // Final boundary prefix
	buffer    *bytes.Buffer // Content waiting to be read
}

// newBoundaryReader returns an initialized boundaryReader
func newBoundaryReader(reader *bufio.Reader, boundary string) *boundaryReader {
	fullBoundary := []byte("\n--" + boundary + "--")
	return &boundaryReader{
		r:        reader,
		nlPrefix: fullBoundary[:len(fullBoundary)-2],
		prefix:   fullBoundary[1 : len(fullBoundary)-2],
		final:    fullBoundary[1:],
		buffer:   new(bytes.Buffer),
	}
}

// readUntilBoundary returns a buffer containing the content up until boundary
func (b *boundaryReader) Read(dest []byte) (n int, err error) {
	if b.buffer.Len() >= len(dest) {
		// This read request can be satisfied entirely by the buffer
		return b.buffer.Read(dest)
	}

	peek, err := b.r.Peek(peekBufferSize)
	peekEOF := (err == io.EOF)
	if err != nil && !peekEOF && err != bufio.ErrBufferFull {
		// Unexpected error
		return 0, err
	}
	var nCopy int
	idx, complete := locateBoundary(peek, b.nlPrefix)
	if idx != -1 {
		// Peeked boundary prefix, read until that point
		nCopy = idx
		if !complete && nCopy == 0 {
			// Incomplete boundary, move past it
			nCopy = 1
		}
	} else {
		// No boundary found, move forward a safe distance
		if nCopy = len(peek) - len(b.nlPrefix) - 1; nCopy <= 0 {
			nCopy = 0
			if peekEOF {
				// No more peek space remaining and no boundary found
				return 0, io.ErrUnexpectedEOF
			}
		}
	}
	if nCopy > 0 {
		if _, err = io.CopyN(b.buffer, b.r, int64(nCopy)); err != nil {
			return 0, err
		}
	}

	n, err = b.buffer.Read(dest)
	if err == io.EOF && !complete {
		// Only the buffer is empty, not the boundaryReader
		return n, nil
	}
	return
}

// Next moves over the boundary to the next part, returns true if there is another part to be read.
func (b *boundaryReader) Next() (bool, error) {
	if b.finished {
		return false, nil
	}
	if b.partsRead > 0 {
		// Exhaust the current part to prevent errors when moving to the next part
		_, _ = io.Copy(ioutil.Discard, b)
	}
	for {
		line, err := b.r.ReadSlice('\n')
		if err != nil && err != io.EOF {
			return false, err
		}
		if len(line) > 0 && (line[0] == '\r' || line[0] == '\n') {
			// Blank line
			continue
		}
		if b.isTerminator(line) {
			b.finished = true
			return false, nil
		}
		if err != io.EOF && b.isDelimiter(line) {
			// Start of a new part
			b.partsRead++
			return true, nil
		}
		if err == io.EOF {
			return false, io.EOF
		}
		if b.partsRead == 0 {
			// The first part didn't find the starting delimiter, burn off any preamble in front of
			// the boundary
			continue
		}
		b.finished = true
		return false, fmt.Errorf("expecting boundary %q, got %q", string(b.prefix), string(line))
	}
}

// isDelimiter returns true for --BOUNDARY\r\n but not --BOUNDARY--
func (b *boundaryReader) isDelimiter(buf []byte) bool {
	idx := bytes.Index(buf, b.prefix)
	if idx == -1 {
		return false
	}

	// Fast forward to the end of the boundary prefix
	buf = buf[idx+len(b.prefix):]
	buf = bytes.TrimLeft(buf, " \t")
	if len(buf) > 0 {
		if buf[0] == '\r' || buf[0] == '\n' {
			return true
		}
	}

	return false
}

// isTerminator returns true for --BOUNDARY--
func (b *boundaryReader) isTerminator(buf []byte) bool {
	idx := bytes.Index(buf, b.final)
	if idx == -1 {
		return false
	}
	return true
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
	if len(buf) > 1 {
		if buf[0] == '-' && buf[1] == '-' {
			return idx, true
		}
	}
	buf = bytes.TrimLeft(buf, " \t")
	if len(buf) > 0 {
		if buf[0] == '\r' || buf[0] == '\n' {
			return idx, true
		}
	}

	return
}
