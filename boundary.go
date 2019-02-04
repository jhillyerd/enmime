package enmime

import (
	"bufio"
	"bytes"
	stderrors "errors"
	"fmt"
	"io"
	"io/ioutil"
	"unicode"

	"github.com/pkg/errors"
)

// This constant needs to be at least 76 for this package to work correctly.  This is because
// \r\n--separator_of_len_70- would fill the buffer and it wouldn't be safe to consume a single byte
// from it.
const peekBufferSize = 4096

var errNoBoundaryTerminator = stderrors.New("expected boundary not present")

type boundaryReader struct {
	finished        bool          // No parts remain when finished
	partsRead       int           // Number of parts read thus far
	r               *bufio.Reader // Source reader
	nlPrefix        []byte        // NL + MIME boundary prefix
	prefix          []byte        // MIME boundary prefix
	final           []byte        // Final boundary prefix
	buffer          *bytes.Buffer // Content waiting to be read
	crBoundryPrefix bool          // Flag for CR in CRLF + MIME boundary
	unbounded       bool          // Flag to throw errNoBoundaryTerminator
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

// Read returns a buffer containing the content up until boundary
func (b *boundaryReader) Read(dest []byte) (n int, err error) {
	if b.buffer.Len() >= len(dest) {
		// This read request can be satisfied entirely by the buffer
		return b.buffer.Read(dest)
	}

	for i := 0; i < cap(dest); i++ {
		oneByte, err := b.r.Peek(1)
		if err != nil && err != io.EOF {
			return 0, errors.WithStack(err)
		}
		if bytes.Equal(oneByte, []byte{'\n'}) {
			peek, err := b.r.Peek(len(b.nlPrefix) + 2)
			switch err {
			case nil:
				// don't check the boundary until we only have one new line
				if bytes.HasPrefix(peek, []byte("\n\n")) ||
					bytes.HasPrefix(peek, []byte("\n\r")) {
					break
				}
				// check for a match against boundary or terminal boundary
				if b.isDelimiter(peek[1:]) || b.isTerminator(peek[1:]) {
					// if we stashed a carriage return, lets pop that back onto the io.Reader
					if b.crBoundryPrefix {
						err = b.r.UnreadByte()
						if err != nil {
							// this should never happen, fatal error
							return 0, errors.WithStack(err)
						}
						b.crBoundryPrefix = false
					}
					n, err = b.buffer.Read(dest)
					switch err {
					case nil, io.EOF:
						return n, io.EOF
					default:
						return 0, errors.WithStack(err)
					}
				}
			default:
				// got to the end without seeing a boundary
				if err == io.EOF {
					b.unbounded = true
					break
				}
				continue
			}
			// wasn't a boundary, write the stored carriage return
			if b.crBoundryPrefix {
				err = b.buffer.WriteByte(byte('\r'))
				if err != nil {
					return 0, errors.WithStack(err)
				}
				b.crBoundryPrefix = false
			}
		}

		// store this carriage return just in case it begins a boundary
		if bytes.Equal(oneByte, []byte{'\r'}) {
			_, err := b.r.ReadByte()
			if err != nil {
				return 0, errors.WithStack(err)
			}
			b.crBoundryPrefix = true
			continue
		}

		_, err = io.CopyN(b.buffer, b.r, 1)
		if err != nil {
			// EOF is not fatal
			if errors.Cause(err) == io.EOF {
				break
			}
			return 0, err
		}
	}

	n, err = b.buffer.Read(dest)
	return n, err
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
			return false, errors.WithStack(err)
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
			// Intentionally not wrapping with stack
			return false, io.EOF
		}
		if b.partsRead == 0 {
			// The first part didn't find the starting delimiter, burn off any preamble in front of
			// the boundary
			continue
		}
		b.finished = true
		return false, errors.WithMessage(errNoBoundaryTerminator, fmt.Sprintf("expecting boundary %q, got %q", string(b.prefix), string(line)))
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
	if len(buf) > 0 {
		if unicode.IsSpace(rune(buf[0])) {
			return true
		}
	}

	return false
}

// isTerminator returns true for --BOUNDARY--
func (b *boundaryReader) isTerminator(buf []byte) bool {
	idx := bytes.Index(buf, b.final)
	return idx != -1
}
