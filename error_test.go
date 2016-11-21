package enmime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestStringConversion(t *testing.T) {
	e := &Error{
		Name:   "WarnName",
		Detail: "Warn Details",
		Severe: false,
	}

	want := "[W] WarnName: Warn Details"
	got := e.String()
	if got != want {
		t.Error("got:", got, "want:", want)
	}

	e = &Error{
		Name:   "ErrorName",
		Detail: "Error Details",
		Severe: true,
	}

	want = "[E] ErrorName: Error Details"
	got = e.String()
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

func TestWarnings(t *testing.T) {
	// To pass each file below must error one or more times with the specified errorName, and no
	// other errorNames.
	var files = []struct {
		filename string
		perror   errorName
	}{
		{"bad-final-boundary.raw", errorMissingBoundary},
		{"missing-content-type.raw", errorMissingContentType},
		// {"unk-charset-html-only.raw", errorCharsetConversion},
		// {"unk-charset-part.raw", errorCharsetConversion},
	}

	for _, tt := range files {
		// Mail with disposition attachment
		msg := readLowQuality(tt.filename)
		e, err := ReadEnvelope(msg)
		if err != nil {
			t.Fatal("Failed to parse MIME:", err)
		}

		if len(e.Errors) == 0 {
			t.Error("Got 0 warnings, expected at least one for:", tt.filename)
		}

		for _, perr := range e.Errors {
			if perr.Severe {
				t.Errorf("Expected Severe to be false, got true")
			}
			if perr.Name != string(tt.perror) {
				t.Errorf("Got error %q, want %q for: %s", perr.Name, tt.perror, tt.filename)
			}
		}
	}
}

// readMessage is a test utility function to fetch a mail.Message object.
func readLowQuality(filename string) io.Reader {
	// Open test email for parsing
	r, err := os.Open(filepath.Join("testdata", "low-quality", filename))
	if err != nil {
		panic(fmt.Sprintf("Failed to open test data: %v", err))
	}
	return r
}
