package enmime

import (
	"bufio"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestBoundaryEOF(t *testing.T) {
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
		t.Errorf("read %v bytes, want: 0")
	}
}

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
			t.Errorf("readUntilBoundary(%q)\ngot: %q, want: %q", tt.input, got, tt.want)
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
