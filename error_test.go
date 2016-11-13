package enmime

import (
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
