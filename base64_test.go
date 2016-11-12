package enmime

import (
	"bytes"
	"strings"
	"testing"
)

func TestBase64Cleaner(t *testing.T) {
	in := "\tA B\r\nC"
	want := "ABC"
	cleaner := NewBase64Cleaner(strings.NewReader(in))
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(cleaner)

	got := buf.String()
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}
