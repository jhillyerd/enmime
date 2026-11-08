package coding

import (
	"bufio"
	"fmt"
	"io"
	"mime/quotedprintable"
)

// QPCleaner scans quoted printable content for invalid characters and encodes them so that
// Go's quoted-printable decoder does not abort with an error.
type QPCleaner struct {
	in       *bufio.Reader
	overflow []byte
	lineLen  int
}

// MaxQPLineLen is the maximum line length we allow before inserting `=\r\n`.  Prevents buffer
// overflows in mime/quotedprintable.Reader.
const MaxQPLineLen = 1024

var (
	_ io.Reader = &QPCleaner{} // Assert QPCleaner implements io.Reader.

	escapedEquals = []byte("=3D") // QP encoded value of an equals sign.
	lineBreak     = []byte("=\r\n")
)

// NewQPCleaner returns a QPCleaner for the specified reader.
func NewQPCleaner(r io.Reader) *QPCleaner {
	return &QPCleaner{
		in:       bufio.NewReader(r),
		overflow: nil,
		lineLen:  0,
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

	// safeWriteByte outputs a single byte, storing overflow for next read. Updates counters.
	safeWriteByte := func(in byte) {
		if n < destLen {
			dest[n] = in
			n++
		} else {
			qp.overflow = append(qp.overflow, in)
		}
		qp.lineLen++
	}

	// writeBytes outputs multiple bytes, storing overflow for next read. Updates counters.
	writeBytes := func(in []byte) {
		nc := copy(dest[n:], in)
		if nc < len(in) {
			// Stash unwritten bytes into overflow.
			qp.overflow = append(qp.overflow, []byte(in[nc:])...)
		}
		n += nc
		qp.lineLen += len(in)
	}

	// ensureLineLen ensures there is room to write `requested` bytes, preventing a line break being
	// inserted in the middle of the escaped string.  The requested count is in addition to the
	// byte that was already reserved for this loop iteration.
	ensureLineLen := func(requested int) {
		if qp.lineLen+requested >= MaxQPLineLen {
			writeBytes(lineBreak)
			qp.lineLen = 0
		}
	}

	// Loop over bytes in qp.in ByteReader while there is space in dest.
	for n < destLen {
		var b byte
		b, err = qp.in.ReadByte()
		if err != nil {
			return n, err
		}

		if qp.lineLen >= MaxQPLineLen {
			writeBytes(lineBreak)
			qp.lineLen = 0
		}

		switch {
		// Pass valid hex bytes through, otherwise escapes the equals symbol.
		case b == '=':
			ensureLineLen(2)

			hexBytes, peekErr := qp.in.Peek(2)
			if peekErr != nil && peekErr != io.EOF {
				return 0, peekErr
			}
			if validHexBytes(hexBytes) {
				safeWriteByte(b)
				break
			}

			var padding int
			padding, err = qp.softLineBreakPadding()
			if err != nil {
				return 0, err
			}
			if padding > 0 {
				// A decoder deletes the padding anyway. Dropping it here keeps the line break we may
				// insert for the line length from landing between the padding and the line break.
				_, _ = qp.in.Discard(padding)
				safeWriteByte(b)
			} else {
				writeBytes(escapedEquals)
			}

		// Valid special character.
		case b == '\t':
			safeWriteByte(b)

		// Valid special characters that reset line length.
		case b == '\r' || b == '\n':
			safeWriteByte(b)
			qp.lineLen = 0

		// Invalid characters, render as quoted-printable.
		case b < ' ' || '~' < b:
			ensureLineLen(2)
			writeBytes(fmt.Appendf(nil, "=%02X", b))

		// Acceptable characters.
		default:
			safeWriteByte(b)
		}
	}

	return n, err
}

// softLineBreakPadding returns the number of spaces and tabs between the equals sign just read
// and a line break. RFC 2045 section 6.7 calls such whitespace transport padding, which a decoder
// deletes, so the equals sign is a soft line break. It returns 0 if the equals sign is not followed
// by padding and a line break, and an error if the lookahead fails with an error other than EOF.
func (qp *QPCleaner) softLineBreakPadding() (int, error) {
	for i := 1; ; i++ {
		peeked, err := qp.in.Peek(i)
		if err != nil {
			return 0, peekError(err)
		}
		switch peeked[i-1] {
		case ' ', '\t':
			continue
		case '\n':
			return i - 1, nil
		case '\r':
			next, err := qp.in.Peek(i + 1)
			if err != nil {
				return 0, peekError(err)
			}
			if next[i] == '\n' {
				return i - 1, nil
			}
			return 0, nil
		default:
			return 0, nil
		}
	}
}

// peekError returns the error of a lookahead, or nil if the input ended or the lookahead did not
// fit into the buffer, in which case the bytes are not a soft line break.
func peekError(err error) error {
	if err == io.EOF || err == bufio.ErrBufferFull {
		return nil
	}
	return err
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

// WriteQuotedPrintable writes content to w as quoted-printable encoded data. No line of the
// encoded data begins with the multipart boundary delimiter "--" + boundary, because a parser
// reads such a line as the end of the part (RFC 2046 section 5.1.1). An empty boundary (the
// content is not part of a multipart) disables the check.
//
// When the encoder would start a line with the delimiter, that line is ended after its first
// hyphen instead, so the rest of the delimiter starts the next one. Decoders remove the soft
// line break this adds, so the decoded content is unchanged. Encoding the first hyphen as =2D
// would have the same effect, but would push the line past the 76 characters of RFC 2045
// section 6.7.
func WriteQuotedPrintable(w io.Writer, content []byte, boundary string) error {
	out := &qpDelimiterWriter{w: w, lineStart: true}
	if boundary != "" {
		out.delim = "--" + boundary
	}

	qp := quotedprintable.NewWriter(out)
	if _, err := qp.Write(content); err != nil {
		return err
	}
	if err := qp.Close(); err != nil {
		return err
	}

	return out.Flush()
}

// qpDelimiterWriter passes quoted-printable encoded data through to an underlying writer, with
// the guarantees of WriteQuotedPrintable.
type qpDelimiterWriter struct {
	w         io.Writer
	delim     string // Delimiter that must not begin a line, empty if there is none.
	buf       []byte // Encoded data for the current Write call.
	hold      []byte // Start of a line that may yet turn out to be the delimiter.
	lineStart bool   // The next byte written begins a line.
}

// Write method for io.Writer interface.
func (qp *qpDelimiterWriter) Write(p []byte) (n int, err error) {
	qp.buf = qp.buf[:0]

	for _, b := range p {
		switch {
		case len(qp.hold) > 0 && b != qp.delim[len(qp.hold)]:
			// Not the delimiter after all, so write out what was held back.
			qp.buf = append(qp.buf, qp.hold...)
			qp.buf = append(qp.buf, b)
			qp.hold = qp.hold[:0]
			qp.lineStart = b == '\n'
		case len(qp.hold) > 0:
			qp.hold = append(qp.hold, b)
			if len(qp.hold) < len(qp.delim) {
				continue
			}
			// The delimiter would begin this line: end the line after its first hyphen.
			qp.buf = append(qp.buf, '-', '=', '\r', '\n')
			qp.buf = append(qp.buf, qp.hold[1:]...)
			qp.hold = qp.hold[:0]
			qp.lineStart = false
		case qp.lineStart && qp.delim != "" && b == '-':
			qp.hold = append(qp.hold[:0], b)
		default:
			qp.buf = append(qp.buf, b)
			qp.lineStart = b == '\n'
		}
	}

	if err := qp.flush(); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Flush writes out any bytes held back while waiting for the rest of a possible delimiter.
func (qp *qpDelimiterWriter) Flush() error {
	if len(qp.hold) > 0 {
		qp.buf = append(qp.buf[:0], qp.hold...)
		qp.hold = qp.hold[:0]
		return qp.flush()
	}
	return nil
}

// flush writes buf to the underlying writer.
func (qp *qpDelimiterWriter) flush() error {
	if len(qp.buf) == 0 {
		return nil
	}
	_, err := qp.w.Write(qp.buf)
	return err
}
