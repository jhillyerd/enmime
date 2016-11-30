package enmime

import "testing"

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
		{"bad-header-wrap.raw", errorMalformedHeader},
		{"html-only-inline.raw", errorPlainTextFromHTML},
		{"missing-content-type.raw", errorMissingContentType},
		{"missing-content-type2.raw", errorMissingContentType},
		{"unk-encoding-part.raw", errorContentEncoding},
		// TODO Implement below cases
		// {"unk-charset-html-only.raw", errorCharsetConversion},
		// {"unk-charset-part.raw", errorCharsetConversion},
	}

	for _, tt := range files {
		// Mail with disposition attachment
		msg := openTestData("low-quality", tt.filename)
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
				t.Errorf(
					"Got error %q, want %q for: %s\n(%s)",
					perr.Name,
					tt.perror,
					tt.filename,
					perr.String())
			}
		}
	}
}
