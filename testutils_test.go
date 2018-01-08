package enmime

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Syntatic sugar for Part comparisons
var partExists = &Part{}

// openTestData is a utility function to open a file in testdata for reading, it will panic if there
// is an error.
func openTestData(subdir, filename string) io.Reader {
	// Open test part for parsing
	raw, err := os.Open(filepath.Join("testdata", subdir, filename))
	if err != nil {
		// err already contains full path to file
		panic(err)
	}
	return raw
}

// comparePart test hehlper compares the contents of two parts, returning true if they are equal.
// t.Errorf() will be called for each field that is not equal.  The presence of child and siblings
// will be checked, but not the contents of them.  Header, Errors and unexported fields are ignored.
func comparePart(t *testing.T, got *Part, want *Part) (equal bool) {
	t.Helper()
	if got == nil && want != nil {
		t.Error("Part == nil, want not nil")
		return
	}
	if got != nil && want == nil {
		t.Error("Part != nil, want nil")
		return
	}
	equal = true
	if got == nil && want == nil {
		return
	}
	if (got.Parent == nil) != (want.Parent == nil) {
		equal = false
		gs := "nil"
		ws := "nil"
		if got.Parent != nil {
			gs = "present"
		}
		if want.Parent != nil {
			ws = "present"
		}
		t.Errorf("Part.Parent == %q, want: %q", gs, ws)
	}
	if (got.FirstChild == nil) != (want.FirstChild == nil) {
		equal = false
		gs := "nil"
		ws := "nil"
		if got.FirstChild != nil {
			gs = "present"
		}
		if want.FirstChild != nil {
			ws = "present"
		}
		t.Errorf("Part.FirstChild == %q, want: %q", gs, ws)
	}
	if (got.NextSibling == nil) != (want.NextSibling == nil) {
		equal = false
		gs := "nil"
		ws := "nil"
		if got.NextSibling != nil {
			gs = "present"
		}
		if want.NextSibling != nil {
			ws = "present"
		}
		t.Errorf("Part.NextSibling == %q, want: %q", gs, ws)
	}
	if got.ContentType != want.ContentType {
		equal = false
		t.Errorf("Part.ContentType == %q, want: %q", got.ContentType, want.ContentType)
	}
	if got.Disposition != want.Disposition {
		equal = false
		t.Errorf("Part.Disposition == %q, want: %q", got.Disposition, want.Disposition)
	}
	if got.FileName != want.FileName {
		equal = false
		t.Errorf("Part.FileName == %q, want: %q", got.FileName, want.FileName)
	}
	if got.Charset != want.Charset {
		equal = false
		t.Errorf("Part.Charset == %q, want: %q", got.Charset, want.Charset)
	}
	if got.PartID != want.PartID {
		equal = false
		t.Errorf("Part.PartID == %q, want: %q", got.PartID, want.PartID)
	}

	return
}

// TestHelperComparePartsEqual tests compareParts with equalivent Parts
func TestHelperComparePartsEqual(t *testing.T) {
	testCases := []struct {
		name string
		part *Part
	}{
		{"nil", nil},
		{"empty", &Part{}},
		{"Parent", &Part{Parent: &Part{}}},
		{"FirstChild", &Part{FirstChild: &Part{}}},
		{"NextSibling", &Part{NextSibling: &Part{}}},
		{"ContentType", &Part{ContentType: "such/wow"}},
		{"Disposition", &Part{Disposition: "irritable"}},
		{"FileName", &Part{FileName: "readme.txt"}},
		{"Charset", &Part{Charset: "utf-7.999"}},
		{"PartID", &Part{PartID: "0.1"}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockt := &testing.T{}
			if !comparePart(mockt, tc.part, tc.part) {
				t.Errorf("Got false while comparing a Part %v to itself: %+v", tc.name, tc.part)
			}
			if mockt.Failed() {
				t.Errorf("Helper failed test for %q, should have been successful", tc.name)
			}
		})
	}
}

// TestHelperComparePartsInequal tests compareParts with differing Parts
func TestHelperComparePartsInequal(t *testing.T) {
	testCases := []struct {
		name string
		a, b *Part
	}{
		{
			name: "nil",
			a:    nil,
			b:    &Part{},
		},
		{
			name: "Parent",
			a:    &Part{},
			b:    &Part{Parent: &Part{}},
		},
		{
			name: "FirstChild",
			a:    &Part{},
			b:    &Part{FirstChild: &Part{}},
		},
		{
			name: "NextSibling",
			a:    &Part{},
			b:    &Part{NextSibling: &Part{}},
		},
		{
			name: "ContentType",
			a:    &Part{ContentType: "text/plain"},
			b:    &Part{ContentType: "text/html"},
		},
		{
			name: "Disposition",
			a:    &Part{Disposition: "happy"},
			b:    &Part{Disposition: "sad"},
		},
		{
			name: "FileName",
			a:    &Part{FileName: "foo.gif"},
			b:    &Part{FileName: "bar.jpg"},
		},
		{
			name: "Charset",
			a:    &Part{Charset: "foo"},
			b:    &Part{Charset: "bar"},
		},
		{
			name: "PartID",
			a:    &Part{PartID: "0"},
			b:    &Part{PartID: "1.1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockt := &testing.T{}
			if comparePart(mockt, tc.a, tc.b) {
				t.Errorf(
					"Got true while comparing inequal Parts (%v):\n"+
						"Part A: %+v\nPart B: %+v", tc.name, tc.a, tc.b)
			}
			if tc.name != "" && !mockt.Failed() {
				t.Errorf("Mock test succeeded for %s, should have failed", tc.name)
			}
		})
	}
}

// contentContainsString checks if the provided readers content contains the specified substring
func contentContainsString(b []byte, substr string) (ok bool, err error) {
	got := string(b)
	if strings.Contains(got, substr) {
		return true, nil
	}
	return false, fmt.Errorf("content == %q, should contain: %q", got, substr)
}

// contentEqualsString checks if the provided readers content is the specified string
func contentEqualsString(b []byte, str string) (ok bool, err error) {
	got := string(b)
	if got == str {
		return true, nil
	}
	return false, fmt.Errorf("content == %q, want: %q", got, str)
}

// contentEqualsBytes checks if the provided readers content is the specified []byte
func contentEqualsBytes(b []byte, want []byte) (ok bool, err error) {
	if bytes.Equal(b, want) {
		return true, nil
	}
	return false, fmt.Errorf("content:\n%v, want:\n%v", b, want)
}
