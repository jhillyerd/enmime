package coding_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime/internal/coding"
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

// TestQPCleanerOverflow attempts to confuse the cleaner by issuing a smaller subsequent read
func TestQPCleanerOverflow(t *testing.T) {
	input := bytes.Repeat([]byte("pédagogues =\r\n"), 1000)
	want := bytes.Repeat([]byte("p=C3=A9dagogues =\r\n"), 1000)
	inbuf := bytes.NewBuffer(input)
	qp := coding.NewQPCleaner(inbuf)

	offset := 0
	for len := 1000; len > 0; len -= 100 {
		p := make([]byte, len)
		n, err := qp.Read(p)
		if err != nil {
			t.Fatal(err)
		}
		if n < 1 {
			t.Fatalf("Read(p) = %v, wanted >0", n)
		}
		for i := 0; i < n; i++ {
			if p[i] != want[offset] {
				t.Errorf("p[%v] = %q, want: %q (want[%v])", i, p[i], want[offset], offset)
			}
			offset++
		}
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

var result int

func BenchmarkQPCleaner(b *testing.B) {
	b.StopTimer()
	input := bytes.Repeat([]byte("pédagogues =\r\n"), b.N)
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
