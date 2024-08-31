package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDefaultExtractor(t *testing.T) {
	if newDefaultExtractor() == nil {
		t.Fatal("Extractor instance should not be nil")
	}
}

func TestExtractEmptyFilename(t *testing.T) {
	b := &bytes.Buffer{}
	testExtractor := &extractor{
		errOut: b,
		stdOut: io.Discard,
	}
	exitCode := testExtractor.extract("", "")
	if exitCode != 1 {
		t.Fatal("Exit code should be 1, failed")
	}
	if !strings.Contains(b.String(), "Missing filename argument") {
		t.Fatal("Should not succeed with an empty filename, failed")
	}
}

func TestExtractEmptyOutputDirectory(t *testing.T) {
	b := &bytes.Buffer{}
	workingDirectoryFn := func() (string, error) {
		return "", nil
	}
	testExtractor := &extractor{
		errOut: b,
		stdOut: io.Discard,
		wd:     workingDirectoryFn,
	}
	exitCode := testExtractor.extract("some.file", "")
	if exitCode != 2 {
		t.Fatal("Exit code should be 2, failed")
	}
	if !strings.Contains(b.String(), "Mkdir  failed.") {
		t.Fatal("Should not succeed with an empty output directory, failed")
	}
}

func TestExtractFailedToOpenFile(t *testing.T) {
	b := &bytes.Buffer{}
	testExtractor := &extractor{
		errOut: b,
		stdOut: io.Discard,
		wd:     os.Getwd,
	}
	exitCode := testExtractor.extract("some.file", "")
	if exitCode != 1 {
		t.Fatal("Exit code should be 1, failed")
	}
	if !strings.Contains(b.String(), "Failed to open file:") {
		t.Fatal("File should not exist, failed")
	}
}

func TestExtractFailedToParse(t *testing.T) {
	s := &bytes.Buffer{}
	testExtractor := &extractor{
		errOut: s,
		stdOut: io.Discard,
		wd:     os.Getwd,
	}
	exitCode := testExtractor.extract(filepath.Join("..", "..", "testdata", "mail", "erroneous.raw"), "")
	if exitCode != 1 {
		t.Fatal("Exit code should be 1, failed")
	}
	if !strings.Contains(s.String(), "During enmime.ReadEnvelope") {
		t.Fatal("Should have failed to write the attachment, failed")
	}
}

func TestExtractAttachmentWriteFail(t *testing.T) {
	s := &bytes.Buffer{}
	fw := func(_ string, _ []byte, _ os.FileMode) error {
		return errors.New("AttachmentWriteFail")
	}
	testExtractor := &extractor{
		errOut:    io.Discard,
		stdOut:    s,
		wd:        os.Getwd,
		fileWrite: fw,
	}
	exitCode := testExtractor.extract(filepath.Join("..", "..", "testdata", "mail", "attachment.raw"), "")
	if exitCode != 0 {
		t.Fatal("Exit code should be 0, failed")
	}
	if !strings.HasSuffix(s.String(), "AttachmentWriteFail\n") {
		t.Fatal("Should have failed to write the attachment, failed")
	}
}

func TestExtractSuccess(t *testing.T) {
	b := &bytes.Buffer{}
	attachmentCount := 0
	fw := func(_ string, _ []byte, _ os.FileMode) error {
		attachmentCount++
		return nil
	}
	testExtractor := &extractor{
		errOut:    b,
		stdOut:    io.Discard,
		wd:        os.Getwd,
		fileWrite: fw,
	}
	exitCode := testExtractor.extract(filepath.Join("..", "..", "testdata", "mail", "attachment.raw"), "")
	if exitCode != 0 {
		t.Fatal("Exit code should be 0, failed")
	}
	if attachmentCount < 1 {
		t.Fatal("Should be one attachment, failed")
	}
	if !strings.Contains(b.String(), "Done!") {
		t.Fatal("Should have succeeded, failed")
	}
}
