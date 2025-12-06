package enmime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorStringConversion(t *testing.T) {
	e := &Error{
		Name:   "WarnName",
		Detail: "Warn Details",
		Severe: false,
	}

	want := "[W] WarnName: Warn Details"
	got := e.Error()
	if got != want {
		t.Error("got:", got, "want:", want)
	}

	e = &Error{
		Name:   "ErrorName",
		Detail: "Error Details",
		Severe: true,
	}

	want = "[E] ErrorName: Error Details"
	got = e.Error()
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

func TestErrorAddError(t *testing.T) {
	p := NewPart("text/plain")
	p.addErrorf(ErrorMalformedHeader, "1 %v %q", 2, "three")

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
	p := NewPart("text/plain")
	p.addWarningf(ErrorMalformedHeader, "1 %v %q", 2, "three")

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
		{"incorrect-charset.raw", ErrorCharsetDeclaration},
	}

	for _, tt := range files {
		t.Run(tt.filename, func(t *testing.T) {
			r, _ := os.Open(filepath.Join("testdata", "low-quality", tt.filename))
			e, err := ReadEnvelope(r)
			if err != nil {
				t.Fatalf("Failed to parse MIME: %+v", err)
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
				var errorList strings.Builder
				for _, perr := range e.Errors {
					errorList.WriteString(perr.Error())
					errorList.WriteString("\n")
				}
				t.Errorf(
					"File %q should have error of type %q, got these instead:\n%s",
					tt.filename,
					tt.perror,
					errorList.String())
			}
		})
	}
}

func TestErrorLimitOption(t *testing.T) {
	addThreeErrors := func(parser *Parser) int {
		part := NewPart("text/plain")
		if parser != nil {
			part.parser = parser
		}

		part.addError("test1", "test1")
		part.addError("test2", "test2")
		part.addError("test3", "test3")

		return len(part.Errors)
	}

	var got, want int

	// Check unlimited by default.
	want = 3
	got = addThreeErrors(nil)
	assert.Equal(t, want, got, "expected unlimited errors")

	// Check unlimited by default when providing Parser.
	want = 3
	got = addThreeErrors(NewParser())
	assert.Equal(t, want, got, "expected unlimited errors")

	// Check limit.
	want = 1
	got = addThreeErrors(NewParser(MaxStoredPartErrors(want)))
	assert.Equal(t, want, got, "expected limited errors")

	// Check limit matching count.
	want = 3
	got = addThreeErrors(NewParser(MaxStoredPartErrors(want)))
	assert.Equal(t, want, got, "expected limited errors")
}
