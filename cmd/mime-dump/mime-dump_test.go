package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDefaultDumper(t *testing.T) {
	if newDefaultDumper() == nil {
		t.Fatal("Dumper instance should not be nil")
	}
}

func TestNotEnoughArgs(t *testing.T) {
	b := &bytes.Buffer{}
	d := &dumper{
		errOut: b,
	}
	exitCode := d.dump(nil)
	if exitCode != 1 {
		t.Fatal("Should have returned an exit code of 1, failed")
	}
	if b.String() != "Missing filename argument\n" {
		t.Fatal("Should be missing filename argument, failed")
	}
}

func TestFailedToOpenFile(t *testing.T) {
	b := &bytes.Buffer{}
	d := &dumper{
		errOut: b,
	}
	exitCode := d.dump([]string{"", ""})
	if exitCode != 1 {
		t.Fatal("Should have returned an exit code of 1, failed")
	}
	if !strings.HasPrefix(b.String(), "Failed to open file") {
		t.Fatal("Should have failed to open file, failed")
	}
}

func TestFailedToParseFile(t *testing.T) {
	b := &bytes.Buffer{}
	d := &dumper{
		errOut: b,
	}
	exitCode := d.dump([]string{"", filepath.Join("..", "..", "testdata", "mail", "erroneous.raw")})
	if exitCode != 1 {
		t.Fatal("Should have returned an exit code of 1, failed")
	}
	if !strings.HasPrefix(b.String(), "Failed to read envelope") {
		t.Fatal("Should have failed to parse file, but couldn't find error message")
	}
}

func TestSuccess(t *testing.T) {
	b := &bytes.Buffer{}
	s := &bytes.Buffer{}
	d := &dumper{
		errOut: b,
		stdOut: s,
	}
	exitCode := d.dump([]string{"", filepath.Join("..", "..", "testdata", "mail", "attachment.raw")})
	if exitCode != 0 {
		t.Fatal("Should have returned an exit code of 0, failed")
	}
	if b.Len() > 0 {
		t.Fatal("Should not have produced any errors, failed")
	}
	if s.Len() == 0 {
		t.Fatal("Should have printed markdown document, failed")
	}
}
