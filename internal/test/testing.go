package test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime"
)

// PartExists indicates to ComparePart that this part is expect to exist
var PartExists = &enmime.Part{}

// OpenTestData is a utility function to open a file in testdata for reading, it will panic if there
// is an error.
func OpenTestData(subdir, filename string) io.Reader {
	// Open test part for parsing
	raw, err := os.Open(filepath.Join("testdata", subdir, filename))
	if err != nil {
		// err already contains full path to file
		panic(err)
	}
	return raw
}

// ComparePart test helper compares the attributes of two parts, returning true if they are equal.
// t.Errorf() will be called for each field that is not equal.  The presence of child and siblings
// will be checked, but not the attributes of them.  Header, Errors and unexported fields are
// ignored.
func ComparePart(t *testing.T, got *enmime.Part, want *enmime.Part) (equal bool) {
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

// ContentContainsString checks if the provided readers content contains the specified substring
func ContentContainsString(t *testing.T, b []byte, substr string) {
	t.Helper()
	got := string(b)
	if !strings.Contains(got, substr) {
		t.Errorf("content == %q, should contain: %q", got, substr)
	}
}

// ContentEqualsString checks if the provided readers content is the specified string
func ContentEqualsString(t *testing.T, b []byte, str string) {
	t.Helper()
	got := string(b)
	if got != str {
		t.Errorf("content == %q, want: %q", got, str)
	}
}

// ContentEqualsBytes checks if the provided readers content is the specified []byte
func ContentEqualsBytes(t *testing.T, b []byte, want []byte) {
	t.Helper()
	if !bytes.Equal(b, want) {
		t.Errorf("content:\n%v, want:\n%v", b, want)
	}
}
