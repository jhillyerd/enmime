package enmime

import (
	"bufio"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"testing"
)

func TestStringConversion(t *testing.T) {
	e := &MIMEError{
		Name:   "WarnName",
		Detail: "Warn Details",
		Severe: false,
	}

	want := "[W] WarnName: Warn Details"
	got := e.String()
	if got != want {
		t.Error("got:", got, "want:", want)
	}

	e = &MIMEError{
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
	// To pass each file below must fail 1 or more times with the specified error name, and no error
	// names.
	var files = []struct {
		filename string
		merror   errorName
	}{
		{"bad-final-boundary.raw", errorBoundaryMissing},
	}

	for _, tt := range files {
		// Mail with disposition attachment
		msg := readLowQuality(tt.filename)
		m, err := ParseMIMEBody(msg)
		if err != nil {
			t.Error("Failed to parse MIME:", err)
		}

		if len(m.Errors) == 0 {
			t.Error("Got 0 warnings, expected at least one for:", tt.filename)
		}

		for _, e := range m.Errors {
			if e.Name != string(tt.merror) {
				t.Errorf("Got error %q, want %q for: %s", e.Name, tt.merror, tt.filename)
			}
		}
	}
}

// readMessage is a test utility function to fetch a mail.Message object.
func readLowQuality(filename string) *mail.Message {
	// Open test email for parsing
	raw, err := os.Open(filepath.Join("testdata", "low-quality", filename))
	if err != nil {
		panic(fmt.Sprintf("Failed to open test data: %v", err))
	}

	// Parse email into a mail.Message object like we do
	reader := bufio.NewReader(raw)
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		panic(fmt.Sprintf("Failed to read message: %v", err))
	}

	return msg
}
