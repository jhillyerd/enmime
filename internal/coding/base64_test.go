package coding_test

import (
	"io"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime/internal/coding"
)

func TestBase64Cleaner(t *testing.T) {
	buf := make([]byte, 1024)
	testCases := []struct {
		input, want string
	}{
		{"", ""},
		{"\tA B\r\nC", "ABC"},
		{"XYZ===", "XYZ"},
	}
	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			cleaner := coding.NewBase64Cleaner(strings.NewReader(tc.input))
			n, err := cleaner.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatal(err)
			}
			for _, e := range cleaner.Errors {
				t.Error(e)
			}
			got := string(buf[:n])
			if got != tc.want {
				t.Error("got:", got, "want:", tc.want)
			}
		})
	}
}

// TestBase64CleanerErrors sends invalid characters and tests error messages
func TestBase64CleanerErrors(t *testing.T) {
	buf := make([]byte, 1024)
	testCases := []struct {
		input, want string
	}{
		{"a!", "a"},
		{"@b", "b"},
		{"#c", "c"},
		{"d$d", "dd"},
		{"ee\b", "ee"},
	}
	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			cleaner := coding.NewBase64Cleaner(strings.NewReader(tc.input))
			n, err := cleaner.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatal(err)
			}
			if len(cleaner.Errors) != 1 {
				t.Errorf("got %d Errors, wanted 1", len(cleaner.Errors))
			}
			got := string(buf[:n])
			if got != tc.want {
				t.Error("got:", got, "want:", tc.want)
			}
		})
	}
}
