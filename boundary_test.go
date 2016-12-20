package enmime

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestBoundaryReader(t *testing.T) {
	var ttable = []struct {
		input, boundary, want string
	}{
		{
			input:    "good\r\n--STOPHERE\r\nafter",
			boundary: "STOPHERE",
			want:     "good",
		},
		{
			input:    "good\r\n--STOPHERE--\r\nafter",
			boundary: "STOPHERE",
			want:     "good",
		},
		{
			input:    "good\r\n--STOPHERE \t\r\nafter",
			boundary: "STOPHERE",
			want:     "good",
		},
		{
			input:    "good\r\n--STOPHERE--\t \r\nafter",
			boundary: "STOPHERE",
			want:     "good",
		},
		{
			input:    "good\r\n--STOPHEREA\r\n--STOPHERE--\r\nafter",
			boundary: "STOPHERE",
			want:     "good\r\n--STOPHEREA",
		},
		{
			input:    "good\r\n--STOPHERE-A\r\n--STOPHERE--\r\nafter",
			boundary: "STOPHERE",
			want:     "good\r\n--STOPHERE-A",
		},
		{
			input:    "good\n--STOPHERE\nafter",
			boundary: "STOPHERE",
			want:     "good",
		},
		{
			input:    "good\n--STOPHERE--\nafter",
			boundary: "STOPHERE",
			want:     "good",
		},
		{
			input:    "good\n--STOPHEREA\n--STOPHERE--\nafter",
			boundary: "STOPHERE",
			want:     "good\n--STOPHEREA",
		},
		{
			input:    "good\n--STOPHERE-A\n--STOPHERE--\nafter",
			boundary: "STOPHERE",
			want:     "good\n--STOPHERE-A",
		},
	}

	for _, tt := range ttable {
		ir := bufio.NewReader(strings.NewReader(tt.input))
		br := newBoundaryReader(ir, tt.boundary)
		output, err := ioutil.ReadAll(br)
		if err != nil {
			t.Fatalf("Got error: %v\ninput: %q", err, tt.input)
		}

		// Test the buffered data is correct
		got := string(output)
		if got != tt.want {
			t.Errorf("boundaryReader input: %q\ngot: %q, want: %q", tt.input, got, tt.want)
		}

		// Test the data remaining in reader is correct
		rest, err := ioutil.ReadAll(ir)
		if err != nil {
			t.Fatal(err)
		}
		got = string(rest)
		want := tt.input[len(tt.want):]
		if got != want {
			t.Errorf("Rest of reader:\ngot: %q, want: %q", got, want)
		}
	}
}

func TestBoundaryReaderBuffer(t *testing.T) {
	// Check that Read() can serve accurately from its buffer
	input := "good\r\n--STOPHERE\r\nafter"
	boundary := "STOPHERE"
	want := []byte("good")

	ir := bufio.NewReader(strings.NewReader(input))
	br := newBoundaryReader(ir, boundary)

	d := make([]byte, 1)
	for i, wc := range want {
		n, err := br.Read(d)
		if err != nil {
			t.Fatal("Unexepcted error:", err)
		}
		if n != 1 {
			t.Error("Got", n, "bytes, want 1")
		}
		if d[0] != wc {
			t.Errorf("Got byte[%v] == %v, want: %v", i, d[0], wc)
		}
	}
	_, err := br.Read(d)
	if err != io.EOF {
		t.Error("Got", err, "wanted: EOF")
	}
}

func TestBoundaryReaderEOF(t *testing.T) {
	// Confirm we get an EOF at end
	input := "good\r\n--STOPHERE\r\nafter"
	boundary := "STOPHERE"
	want := "good"

	ir := bufio.NewReader(strings.NewReader(input))
	br := newBoundaryReader(ir, boundary)
	output, err := ioutil.ReadAll(br)
	if err != nil {
		t.Fatal(err)
	}

	got := string(output)
	if got != want {
		t.Fatal("got:", got, "want:", want)
	}

	buf := make([]byte, 256)
	n, err := br.Read(buf)
	if err != io.EOF {
		t.Error("got:", err, "want: EOF")
	}
	if 0 != n {
		t.Error("read ", n, "bytes, want: 0")
	}
}

func TestBoundaryReaderParts(t *testing.T) {
	var ttable = []struct {
		input    string
		boundary string
		parts    []string
	}{
		{
			input:    "preamble\r\n--STOP\r\npart1\r\n--STOP\r\npart2\r\n--STOP--\r\n",
			boundary: "STOP",
			parts:    []string{"part1", "part2"},
		},
		{
			input:    "preamble\r\n--STOP \t\r\npart1\r\n--STOP\t \r\npart2\r\n--STOP-- \t\r\n",
			boundary: "STOP",
			parts:    []string{"part1", "part2"},
		},
		{
			input:    "\npreamble\n--STOP\npart1\n--STOP\npart2\n--STOP--\n",
			boundary: "STOP",
			parts:    []string{"part1", "part2"},
		},
		{
			input:    "\n--STOP\npart1\n--STOP\npart2\n--STOP--\n",
			boundary: "STOP",
			parts:    []string{"part1", "part2"},
		},
		{
			input:    "--STOP\npart1\n--STOP\npart2\n--STOP--\n",
			boundary: "STOP",
			parts:    []string{"part1", "part2"},
		},
	}

	for _, tt := range ttable {
		ir := bufio.NewReader(strings.NewReader(tt.input))
		br := newBoundaryReader(ir, tt.boundary)

		for i, want := range tt.parts {
			next, err := br.Next()
			if err != nil {
				t.Fatalf("Error %q on part %v, input %q", err, i, tt.input)
			}
			if !next {
				t.Fatal("Next() = false, want: true")
			}
			output, err := ioutil.ReadAll(br)
			if err != nil {
				t.Fatal(err)
			}

			got := string(output)
			if got != want {
				t.Errorf("boundaryReader input: %q\ngot: %q, want: %q", tt.input, got, want)
			}
		}

		next, err := br.Next()
		if err != nil {
			t.Fatal(err)
		}
		if next {
			t.Fatal("Next() = true, want: false")
		}

		// How does it handle being called a second time?
		next, err = br.Next()
		if err != nil {
			t.Fatal(err)
		}
		if next {
			t.Fatal("Next() = true, want: false")
		}
	}
}

func TestBoundaryReaderPartialRead(t *testing.T) {
	// Make sure Next() still works after a partial read
	input := "\r\n--STOPHERE\r\n1111\r\n--STOPHERE\r\n2222\r\n--STOPHERE\r\n"
	boundary := "STOPHERE"
	wants := []string{"11", "2222"}

	ir := bufio.NewReader(strings.NewReader(input))
	br := newBoundaryReader(ir, boundary)

	for i, want := range wants {
		next, err := br.Next()
		if err != nil {
			t.Fatalf("Error %q on part %v, input %q", err, i, input)
		}
		if !next {
			t.Fatal("Next() = false, want: true")
		}

		// Build a buffer the size of our wanted string
		b := make([]byte, len(want))
		count, err := br.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if count != len(want) {
			t.Errorf("Read() size = %v, wanted %v", count, len(want))
		}

		got := string(b[:count])
		if got != want {
			t.Errorf("boundaryReader got: %q, want: %q", got, want)
		}
	}
}

func TestBoundaryReaderNoMatch(t *testing.T) {
	input := "\r\n--STOPHERE\r\n1111\r\n--STOPHERE\r\n2222\r\n--STOPHERE\r\n"
	boundary := "NOMATCH"

	ir := bufio.NewReader(strings.NewReader(input))
	br := newBoundaryReader(ir, boundary)

	next, err := br.Next()
	if err != io.EOF {
		t.Fatalf("err = %v, want: io.EOF", err)
	}
	if next {
		t.Fatalf("Next() = true, want: false")
	}
}

func TestBoundaryReaderNoTerminator(t *testing.T) {
	input := "preamble\r\n--STOPHERE\r\n1111\r\n"
	boundary := "STOPHERE"

	ir := bufio.NewReader(strings.NewReader(input))
	br := newBoundaryReader(ir, boundary)

	// First part should not error
	next, err := br.Next()
	if err != nil {
		t.Fatalf("Error %q on first part, input %q", err, input)
	}
	if !next {
		t.Fatal("Next() = false, want: true")
	}

	// Second part should error
	want := "expecting boundary"
	next, err = br.Next()
	if err == nil {
		t.Fatal("Error was nil, wanted:", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("err = %v, want: %v", err, want)
	}
	if next {
		t.Fatalf("Next() = true, want: false")
	}
}

func TestBoundaryReaderBufferBoundaryAbut(t *testing.T) {
	// Verify operation when the boundary string abuts the end of the peek buffer
	prefix := []byte("preamble\r\n--STOPHERE\r\n")
	peekSuffix := []byte("\r\n--STOPHERE")
	afterPeek := []byte("\r\nanother part\r\n--STOPHERE--")
	buf := make([]byte, 0, len(prefix)+peekBufferSize+len(afterPeek))
	boundary := "STOPHERE"

	// Setup buffer
	buf = append(buf, prefix...)
	padding := peekBufferSize - len(peekSuffix)
	for i := 0; i < padding; i++ {
		buf = append(buf, 'x')
	}
	buf = append(buf, peekSuffix...)
	buf = append(buf, afterPeek...)

	// Attempt to read
	ir := bufio.NewReader(bytes.NewBuffer(buf))
	br := newBoundaryReader(ir, boundary)

	// Skip preamble, first part should not error
	next, err := br.Next()
	if err != nil {
		t.Fatalf("Error %q on first part", err)
	}
	if !next {
		t.Fatal("Next() = false, want: true")
	}
	output, err := ioutil.ReadAll(br)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	if len(output) != padding {
		t.Errorf("len(output) == %v, want %v", len(output), padding)
	}

	// Second part should not error
	next, err = br.Next()
	if err != nil {
		t.Fatalf("Error %q on second part", err)
	}
	if !next {
		t.Fatal("Next() = false, want: true")
	}
	output, err = ioutil.ReadAll(br)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	want := "another part"
	got := string(output)
	if got != want {
		t.Errorf("ReadAll() got: %q, want: %q", got, want)
	}
}

func TestBoundaryReaderBufferBoundaryCross(t *testing.T) {
	// Verify operation when the boundary string does not fit in the peek buffer
	prefix := []byte("preamble\r\n--STOPHERE\r\n")
	peekSuffix := []byte("\r\n--STOP")
	afterPeek := []byte("HERE\r\nanother part\r\n--STOPHERE--")
	buf := make([]byte, 0, len(prefix)+peekBufferSize+len(afterPeek))
	boundary := "STOPHERE"

	// Setup buffer
	buf = append(buf, prefix...)
	padding := peekBufferSize - len(peekSuffix)
	for i := 0; i < padding; i++ {
		buf = append(buf, 'x')
	}
	buf = append(buf, peekSuffix...)
	buf = append(buf, afterPeek...)

	// Attempt to read
	ir := bufio.NewReader(bytes.NewBuffer(buf))
	br := newBoundaryReader(ir, boundary)

	// Skip preamble, first part should not error
	next, err := br.Next()
	if err != nil {
		t.Fatalf("Error %q on first part", err)
	}
	if !next {
		t.Fatal("Next() = false, want: true")
	}
	output, err := ioutil.ReadAll(br)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	if len(output) != padding {
		t.Errorf("len(output) == %v, want %v", len(output), padding)
	}

	// Second part should not error
	next, err = br.Next()
	if err != nil {
		t.Fatalf("Error %q on second part", err)
	}
	if !next {
		t.Fatal("Next() = false, want: true")
	}
	output, err = ioutil.ReadAll(br)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	want := "another part"
	got := string(output)
	if got != want {
		t.Errorf("ReadAll() got: %q, want: %q", got, want)
	}
}
