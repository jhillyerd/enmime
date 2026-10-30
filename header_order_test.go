package enmime_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime/v2"
	"github.com/jhillyerd/enmime/v2/internal/test"
)

func TestPartEncodePreservesHeaderOrder(t *testing.T) {
	r := test.OpenTestData("encode", "parser-raw-content-html-option.raw")
	root, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := root.Encode(&buf); err != nil {
		t.Fatal(err)
	}

	wantPrefix := strings.Join([]string{
		"Subject: Test subject",
		"From: me@me.com",
		"To: you@you.com",
		"Mime-Version: 1.0",
		"Content-Type: text/html; charset=utf-8",
		"",
	}, "\r\n")
	if !strings.HasPrefix(buf.String(), wantPrefix) {
		t.Fatalf("encoded headers:\n%s\nwant prefix:\n%s", buf.String(), wantPrefix)
	}
}
