package enmime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestErrorStringConversion(t *testing.T) {
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

func TestErrorAddError(t *testing.T) {
	p := &Part{}
	p.addError(ErrorMalformedHeader, "1 %v %q", 2, "three")

	if len(p.Errors) != 1 {
		t.Fatal("len(p.Errors) ==", len(p.Errors), ", want: 1")
	}
	e := p.Errors[0]

	if e.Name != ErrorMalformedHeader {
		t.Errorf("e.Name == %q, want: %q", e.Name, ErrorMalformedHeader)
	}
	if !e.Severe {
		t.Errorf("e.Severe == %v, want: true", e.Severe)
	}
	want := "1 2 \"three\""
	if e.Detail != want {
		t.Errorf("e.Detail == %q, want: %q", e.Detail, want)
	}
}

func TestErrorAddWarning(t *testing.T) {
	p := &Part{}
	p.addWarning(ErrorMalformedHeader, "1 %v %q", 2, "three")

	if len(p.Errors) != 1 {
		t.Fatal("len(p.Errors) ==", len(p.Errors), ", want: 1")
	}
	e := p.Errors[0]

	if e.Name != ErrorMalformedHeader {
		t.Errorf("e.Name == %q, want: %q", e.Name, ErrorMalformedHeader)
	}
	if e.Severe {
		t.Errorf("e.Severe == %v, want: false", e.Severe)
	}
	want := "1 2 \"three\""
	if e.Detail != want {
		t.Errorf("e.Detail == %q, want: %q", e.Detail, want)
	}
}

func TestErrorEnvelopeWarnings(t *testing.T) {
	// To pass each file below must error one or more times with the specified errorName, and no
	// other errorNames.
	var files = []struct {
		filename string
		perror   string
	}{
		{"bad-final-boundary.raw", ErrorMissingBoundary},
		{"bad-header-wrap.raw", ErrorMalformedHeader},
		{"html-only-inline.raw", ErrorPlainTextFromHTML},
		{"missing-content-type2.raw", ErrorMissingContentType},
		{"empty-header.raw", ErrorMissingContentType},
		{"unk-encoding-part.raw", ErrorContentEncoding},
		{"unk-charset-html-only.raw", ErrorCharsetConversion},
		{"unk-charset-part.raw", ErrorCharsetConversion},
		{"malformed-base64-attach.raw", ErrorMalformedBase64},
	}

	for _, tt := range files {
		t.Run(tt.filename, func(t *testing.T) {
			r, _ := os.Open(filepath.Join("testdata", "low-quality", tt.filename))
			e, err := ReadEnvelope(r)
			if err != nil {
				t.Fatal("Failed to parse MIME:", err)
			}

			if len(e.Errors) == 0 {
				t.Error("Got 0 warnings, expected at least one for:", tt.filename)
			}

			satisfied := false
			for _, perr := range e.Errors {
				if perr.Name == tt.perror {
					satisfied = true
					if perr.Severe {
						t.Errorf("Expected Severe to be false, got true for %q", perr.Name)
					}
				}
			}
			if !satisfied {
				var errorList string
				for _, perr := range e.Errors {
					errorList += perr.String()
					errorList += "\n"
				}
				t.Errorf(
					"File %q should have error of type %q, got these instead:\n%s",
					tt.filename,
					tt.perror,
					errorList)
			}
		})
	}
}
