package enmime

import (
	"bufio"
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
			t.Fatal(err)
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
