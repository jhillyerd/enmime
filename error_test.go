package enmime

import (
	"os"
	"path/filepath"
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

	// Using deprecated String() method here for test coverage
	want = "[E] ErrorName: Error Details"
	got = e.String()
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

func TestErrorAddError(t *testing.T) {
	p := NewPart("text/plain")
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
	p := NewPart("text/plain")
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
				var errorList string
				for _, perr := range e.Errors {
					errorList += perr.Error()
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

// Deprecated.
func TestErrorLimit(t *testing.T) {
	// Backup global variable
	originalMaxPartErros := MaxPartErrors
	defer func() {
		MaxPartErrors = originalMaxPartErros
	}()

	addThreeErrors := func() int {
		part := NewPart("text/plain")
		part.addError("test1", "test1")
		part.addError("test2", "test2")
		part.addError("test3", "test3")

		return len(part.Errors)
	}

	// Check unlimited
	var errCount int
	MaxPartErrors = 0
	errCount = addThreeErrors()
	if errCount != 3 {
		t.Errorf("Expected unlimited errors (3), got %d", errCount)
	}

	// Check limit
	MaxPartErrors = 1
	errCount = addThreeErrors()
	if errCount != 1 {
		t.Errorf("Expected limited errors (1), got %d", errCount)
	}

	// Check limit matching count
	MaxPartErrors = 3
	errCount = addThreeErrors()
	if errCount != 3 {
		t.Errorf("Expected limited errors (3), got %d", errCount)
	}
}

func TestErrorLimitOption(t *testing.T) {
	// Backup global variable
	originalMaxPartErros := MaxPartErrors
	defer func() {
		MaxPartErrors = originalMaxPartErros
	}()

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

	// Check the default actually comes from deprecated MaxPartErrors global.
	want = 1
	MaxPartErrors = want
	got = addThreeErrors(nil)
	assert.Equal(t, want, got, "expected limited errors")
	MaxPartErrors = 0

	// Check limit.
	want = 1
	got = addThreeErrors(NewParser(MaxStoredPartErrors(want)))
	assert.Equal(t, want, got, "expected limited errors")

	// Check limit matching count.
	want = 3
	got = addThreeErrors(NewParser(MaxStoredPartErrors(want)))
	assert.Equal(t, want, got, "expected limited errors")
}
