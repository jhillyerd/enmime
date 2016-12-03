package enmime

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Syntatic sugar for Part comparisons
var partExists = &Part{}

// inequalPartField is called by comparePart when it finds inequal Part fields
type inequalPartField func(field, got, want string)

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

// comparePart compares the contents of two parts, returning true if they are equal.  The provided
// function will be call for each field that is not equal.  The presence of child and siblings will
// be checked, but not the contents of them.  Header, Errors and unexported fields are ignored.
func comparePart(got *Part, want *Part, handler inequalPartField) (equal bool) {
	if got == nil {
		return want == nil
	}
	equal = true
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
		handler("Parent", gs, ws)
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
		handler("FirstChild", gs, ws)
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
		handler("NextSibling", gs, ws)
	}
	if got.ContentType != want.ContentType {
		equal = false
		handler("ContentType", got.ContentType, want.ContentType)
	}
	if got.Disposition != want.Disposition {
		equal = false
		handler("Disposition", got.Disposition, want.Disposition)
	}
	if got.FileName != want.FileName {
		equal = false
		handler("FileName", got.FileName, want.FileName)
	}
	if got.Charset != want.Charset {
		equal = false
		handler("Charset", got.Charset, want.Charset)
	}
	return
}

func TestTestComparePartsEqual(t *testing.T) {
	// Test equal Parts
	var ttable = []*Part{
		nil,
		&Part{},
		&Part{Parent: &Part{}},
		&Part{FirstChild: &Part{}},
		&Part{NextSibling: &Part{}},
		&Part{ContentType: "such/wow"},
		&Part{Disposition: "irritable"},
		&Part{FileName: "readme.txt"},
		&Part{Charset: "utf-7.999"},
	}

	for i, tt := range ttable {
		if !comparePart(tt, tt, func(field, got, want string) {
			t.Errorf("inequalPartField handler was called for %q, should not have been", field)
		}) {
			t.Errorf("Got false while comparing a Part (index %v) to itself: %+v", i, tt)
		}
	}
}

func TestTestComparePartsInequal(t *testing.T) {
	// Test equal Parts
	var ttable = []struct {
		got, want                  *Part
		field, gotValue, wantValue string
	}{
		{
			got:  nil,
			want: &Part{},
		},
		{
			got:       &Part{},
			want:      &Part{Parent: &Part{}},
			field:     "Parent",
			gotValue:  "nil",
			wantValue: "present",
		},
		{
			got:       &Part{},
			want:      &Part{FirstChild: &Part{}},
			field:     "FirstChild",
			gotValue:  "nil",
			wantValue: "present",
		},
		{
			got:       &Part{},
			want:      &Part{NextSibling: &Part{}},
			field:     "NextSibling",
			gotValue:  "nil",
			wantValue: "present",
		},
		{
			got:       &Part{ContentType: "text/plain"},
			want:      &Part{ContentType: "text/html"},
			field:     "ContentType",
			gotValue:  "text/plain",
			wantValue: "text/html",
		},
		{
			got:       &Part{Disposition: "happy"},
			want:      &Part{Disposition: "sad"},
			field:     "Disposition",
			gotValue:  "happy",
			wantValue: "sad",
		},
		{
			got:       &Part{FileName: "foo.gif"},
			want:      &Part{FileName: "bar.jpg"},
			field:     "FileName",
			gotValue:  "foo.gif",
			wantValue: "bar.jpg",
		},
		{
			got:       &Part{Charset: "foo"},
			want:      &Part{Charset: "bar"},
			field:     "Charset",
			gotValue:  "foo",
			wantValue: "bar",
		},
	}

	for i, tt := range ttable {
		called := false
		var gotField, gotValue, wantValue string
		if comparePart(tt.got, tt.want, func(field, got, want string) {
			called = true
			gotField = field
			gotValue = got
			wantValue = want
		}) {
			t.Errorf(
				"Got true while comparing inequal Parts (index %v, field %v):\n"+
					"got arg: %+v\nwant arg: %+v", i, tt.field, tt.got, tt.want)
		}
		if tt.field != "" {
			if called {
				// Build up handler args as a string: easy to compare and print
				wantArgs := fmt.Sprintf("%q, %q, %q", tt.field, tt.gotValue, tt.wantValue)
				gotArgs := fmt.Sprintf("%q, %q, %q", gotField, gotValue, wantValue)
				if gotArgs != wantArgs {
					t.Errorf(
						"Got: inequalPartField(%s), want: inequalPartField(%s)",
						gotArgs,
						wantArgs)
				}
			} else {
				t.Error("Expected inequalPartField handler to be called, was not for index", i)
			}
		}
	}
}

// contentContainsString checks if the provided readers content contains the specified substring
func contentContainsString(r io.Reader, substr string) (ok bool, err error) {
	allBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return false, err
	}
	got := string(allBytes)
	if strings.Contains(got, substr) {
		return true, nil
	}
	return false, fmt.Errorf("content == %q, should contain: %q", got, substr)
}

// contentEqualsString checks if the provided readers content is the specified string
func contentEqualsString(r io.Reader, str string) (ok bool, err error) {
	allBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return false, err
	}
	got := string(allBytes)
	if got == str {
		return true, nil
	}
	return false, fmt.Errorf("content == %q, want: %q", got, str)
}
