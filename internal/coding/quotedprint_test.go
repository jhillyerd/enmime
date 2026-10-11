package coding_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime/v2/internal/coding"
)

func TestQPCleaner(t *testing.T) {
	ttable := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"abcDEF_", "abcDEF_"},
		{"=5bSlack=5d", "=5bSlack=5d"},
		{"low: ,high:~", "low: ,high:~"},
		{"\r\n\t", "\r\n\t"},
		{"pédagogues", "p=C3=A9dagogues"},
		{"Stuffs’s", "Stuffs=E2=80=99s"},
		{"=", "=3D"},
		{"=a", "=3Da"},
	}

	for _, tc := range ttable {
		// Run cleaner
		cleaner := coding.NewQPCleaner(strings.NewReader(tc.input))
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(cleaner)
		if err != nil {
			t.Fatal(err)
		}

		got := buf.String()
		if got != tc.want {
			t.Errorf("Got: %q, want: %q", got, tc.want)
		}
	}
}

// TestQPCleanerOverflow attempts to confuse the cleaner by issuing smaller subsequent reads.
func TestQPCleanerOverflow(t *testing.T) {
	input := bytes.Repeat([]byte("pédagogues =\r\n"), 1000)
	want := bytes.Repeat([]byte("p=C3=A9dagogues =\r\n"), 1000)
	inbuf := bytes.NewBuffer(input)
	qp := coding.NewQPCleaner(inbuf)

	offset := 0
	for size := 1000; size > 0; size -= 100 {
		buf := make([]byte, size)
		n, err := qp.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if n < 1 {
			t.Fatalf("Read(p) = %v, wanted >0, at want[%v]", n, offset)
		}
		for i := range n {
			if buf[i] != want[offset] {
				t.Errorf("p[%v] = %q, want: %q (want[%v])", i, buf[i], want[offset], offset)
			}
			offset++
		}
	}
}

// TestQPCleanerSmallDest repeatedly calls Read with a small destination buffer.
func TestQPCleanerSmallDest(t *testing.T) {
	input := bytes.Repeat([]byte("pédagogues =z =\r\n"), 100)
	want := bytes.Repeat([]byte("p=C3=A9dagogues =3Dz =\r\n"), 100)

	for bufSize := 5; bufSize > 0; bufSize-- {
		t.Run(fmt.Sprintf("%v byte buffer", bufSize), func(t *testing.T) {
			inbuf := bytes.NewBuffer(input)
			qp := coding.NewQPCleaner(inbuf)

			offset := 0
			p := make([]byte, bufSize)
			for {
				n, err := qp.Read(p)
				if err != nil && err != io.EOF {
					t.Fatal(err)
				}
				if n < 1 && offset < len(want) {
					t.Fatalf("Read(p) = %v, wanted >0, at want[%v]", n, offset)
				}
				for i := range n {
					if p[i] != want[offset] {
						t.Errorf("p[%v] = %q, want: %q (want[%v])", i, p[i], want[offset], offset)
					}
					offset++
				}
				if err == io.EOF {
					break
				}
			}
		})
	}
}

// TestQPCleanerLineBreak verifies QPCleaner breaks long lines correctly.
func TestQPCleanerLineBreak(t *testing.T) {
	input := bytes.Repeat([]byte("pédagogues =z "), 10000)
	inbuf := bytes.NewBuffer(input)
	qp := coding.NewQPCleaner(inbuf)

	output, err := io.ReadAll(qp)
	if err != nil {
		t.Fatal(err)
	}

	want := 1024 // Desired wrapping point.
	tolerance := 3

	if len(output) < want {
		t.Fatalf("wanted minimum output len %v, got %v", want, len(output))
	}

	// Examine each line of output long enough to wrap.
	for i := 0; len(output) > want; i++ {
		got := bytes.Index(output, []byte("=\r\n"))
		// Wrapping a few characters early is OK, but not late.
		if got > want || want-got > tolerance {
			t.Errorf("iteration %v: got line break at %v, wanted %v +/- %v",
				i, got, want, tolerance)
		}
		if got == 0 {
			break
		}
		output = output[got+3:] // Extend past =\r\n
	}
}

func TestQPCleanerLineBreakBufferFull(t *testing.T) {
	prefix := strings.Repeat("a", coding.MaxQPLineLen)
	testCases := []struct {
		name, input, want string
	}{
		{"plain", prefix + "END", prefix + "=\r\nEND"},
		{"tab", prefix + "\tEND", prefix + "=\r\n\tEND"},
		{"newline", prefix + "\r\nEND", prefix + "=\r\n\r\nEND"},
		{"hex-escape", prefix + "=3DEND", prefix + "=\r\n=3DEND"},
		{"invalid-equals", prefix + "=zEND", prefix + "=\r\n=3DzEND"},
		{"trailing-equals", prefix + "=", prefix + "=\r\n=3D"},
		{"trailing-partial-escape", prefix + "=a", prefix + "=\r\n=3Da"},
		{"short-trailing-equals", "=", "=3D"},
		{"short-trailing-partial-escape", "=a", "=3Da"},
		{"invalid-byte", prefix + "\x80END", prefix + "=\r\n=80END"},
		{"eof", prefix, prefix},
	}
	bufferSizes := [][]int{
		{1}, {2}, {3}, {1024}, {1025}, {1026}, {1027}, {1028},
		{1, 1025, 2, 1026, 3, 1027},
	}
	for _, tc := range testCases {
		for _, sizes := range bufferSizes {
			t.Run(fmt.Sprintf("%s/%v", tc.name, sizes), func(t *testing.T) {
				qp := coding.NewQPCleaner(strings.NewReader(tc.input))
				var output strings.Builder
				for reads := 0; ; reads++ {
					dest := make([]byte, sizes[reads%len(sizes)])
					n, err := qp.Read(dest)
					if err != nil && err != io.EOF {
						t.Fatal(err)
					}
					output.Write(dest[:n])
					if err == io.EOF {
						break
					}
				}
				if got, want := output.String(), tc.want; got != want {
					t.Errorf("got %d bytes ending %q, want %d bytes ending %q",
						len(got), got[max(0, len(got)-12):], len(want), want[max(0, len(want)-12):])
				}
			})
		}
	}
}

func TestQPCleanerEqualSignOverflow(t *testing.T) {
	// Test processing of =3D quoted printable token being at the end of line
	// in combination with almost full buffer.
	// qp cleaner should add =\r\n and then output =3D without buffer overflow

	input := append(bytes.Repeat([]byte("abc"), 341), []byte("=3D")...)
	inbuf := bytes.NewBuffer(input)
	qp := coding.NewQPCleaner(inbuf)

	output := make([]byte, 1024)
	n, err := qp.Read(output)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1024 {
		t.Error("got:", n, "want:", 1024)
	}
	if string(output[1020:]) != "abc=" {
		t.Error("got:", string(output[1020:]), "want:", "abc=")
	}

	n, err = qp.Read(output)
	if err != io.EOF {
		t.Fatal(err)
	}
	if n != 5 {
		t.Error("got:", n, "want:", 5)
	}
	output = output[:n]
	if string(output) != "\r\n=3D" {
		t.Error("got:", string(output), "want:", "\r\n=3D")
	}
}

var ErrPeek = errors.New("enmime test peek error")

type peekBreakReader string

// Read always returns a ErrPeek
func (r peekBreakReader) Read(p []byte) (int, error) {
	return copy(p, r), ErrPeek
}

func TestQPPeekError(t *testing.T) {
	qp := coding.NewQPCleaner(peekBreakReader("=a"))

	buf := make([]byte, 100)
	_, err := qp.Read(buf)
	if err != ErrPeek {
		t.Errorf("Got: %q, want: %q", err, ErrPeek)
	}
}

func TestQPCleanerQuotedLineLength(t *testing.T) {
	input := strings.Repeat("=BC", 700) // ~ two lines of token
	inr := strings.NewReader(input)
	qp := coding.NewQPCleaner(inr)

	// Check line length is counted proerly even for quoted printable encoded chars
	longLineLen := coding.MaxQPLineLen + 2
	output := make([]byte, longLineLen)
	n, err := qp.Read(output)
	if err != nil {
		t.Fatal(err)
	}
	if n != longLineLen {
		t.Error("got:", n, "want:", longLineLen)
	}
	if string(output[longLineLen-2:]) != "\r\n" {
		t.Error("got:", string(output[longLineLen-2:]), "want:", "\r\n")
	}

	// Check line length is correct also when overflow buffer of QPCleaner is used
	output = make([]byte, longLineLen)
	n, err = qp.Read(output)
	if err != nil {
		t.Fatal(err)
	}
	if n != longLineLen {
		t.Error("got:", n, "want:", longLineLen)
	}
	if string(output[longLineLen-2:]) != "\r\n" {
		t.Error("got:", string(output[longLineLen-2:]), "want:", "\r\n")
	}
}

var result int

func BenchmarkQPCleaner(b *testing.B) {
	b.StopTimer()
	input := bytes.Repeat([]byte("pédagogues\t =zz =\r\n"), b.N)
	b.SetBytes(int64(len(input)))
	inbuf := bytes.NewBuffer(input)
	qp := coding.NewQPCleaner(inbuf)
	p := make([]byte, 1024)
	b.StartTimer()

	for {
		n, err := qp.Read(p)
		result += n
		if err == io.EOF {
			break
		}
		if err != nil {
			b.Fatalf("Read(): %v", err)
		}
	}
}
