package enmime

import (
	"bufio"
	"fmt"
	"github.com/stretchrcom/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestPlainTextPart(t *testing.T) {
	r := openPart("textplain.raw")
	p, err := ParseMIME(r)

	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, p.Header.Get("Content-Transfer-Encoding"), "7bit",
		"Exepcted Header to have data")
	assert.Equal(t, p.Type, "text/plain", "Expected type to be set")
	assert.Contains(t, string(p.Content), "Test of text/plain section",
		"Expected correct data in p.Content")
	assert.Nil(t, p.NextSibling, "Root should never have a sibling")
}

func TestQuotedPrintablePart(t *testing.T) {
	r := openPart("quoted-printable.raw")
	p, err := ParseMIME(r)

	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, p.Header.Get("Content-Transfer-Encoding"), "quoted-printable",
		"Exepcted Header to have data")
	assert.Equal(t, p.Type, "text/plain", "Expected type to be set")
	assert.Equal(t, string(p.Content), "Start=ABC=Finish",
		"Expected correct data in p.Content")
	assert.Nil(t, p.NextSibling, "Root should never have a sibling")
}

func TestMultiAlternParts(t *testing.T) {
	r := openPart("multialtern.raw")
	p, err := ParseMIME(r)

	// Examine root
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, p.Type, "multipart/alternative", "Expected type to be set")
	assert.Equal(t, len(p.Content), 0, "Root should not have Content")
	assert.NotNil(t, p.FirstChild, "Root should have a FirstChild")
	assert.Nil(t, p.NextSibling, "Root should never have a sibling")

	// Examine first child
	p = p.FirstChild
	assert.Equal(t, p.Type, "text/plain", "First child should have been text")
	assert.Contains(t, string(p.Content), "A text section", "First child contains wrong content")
	assert.NotNil(t, p.NextSibling, "First child should have a sibling")

	// Examine sibling
	p = p.NextSibling
	assert.Equal(t, p.Type, "text/html", "Second child should have been html")
	assert.Contains(t, string(p.Content), "An HTML section", "Second child contains wrong content")
	assert.Nil(t, p.NextSibling, "Second child should not have a sibling")
}

func TestNestedAlternParts(t *testing.T) {
	r := openPart("nestedmulti.raw")
	p, err := ParseMIME(r)

	// Examine root
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, p.Type, "multipart/alternative", "Expected type to be set")
	assert.Equal(t, len(p.Content), 0, "Root should not have Content")
	assert.NotNil(t, p.FirstChild, "Root should have a FirstChild")
	assert.Nil(t, p.NextSibling, "Root should never have a sibling")

	// Examine first child
	p = p.FirstChild
	assert.Equal(t, p.Type, "text/plain", "First child should have been text")
	assert.Contains(t, string(p.Content), "A text section", "First child contains wrong content")
	assert.NotNil(t, p.NextSibling, "First child should have a sibling")

	// Examine sibling
	p = p.NextSibling
	assert.Equal(t, p.Type, "multipart/related", "Second child should have been another multipart")
	assert.Equal(t, len(p.Content), 0, "Second child should not have Content")
	assert.Nil(t, p.NextSibling, "Second child should not have a sibling")
	assert.NotNil(t, p.FirstChild, "Second child should have a child")

	// First nested
	p = p.FirstChild
	assert.Equal(t, p.Type, "text/html", "First nested should have been html")
	assert.Contains(t, string(p.Content), "An HTML section", "First nested contains wrong content")
	assert.NotNil(t, p.NextSibling, "First nested should have a sibling")

	// Second nested
	p = p.NextSibling
	assert.Equal(t, p.Type, "text/plain", "Second nested should have been text")
	assert.Equal(t, p.Disposition, "inline", "Second nested should be inline disposition")
	assert.Equal(t, p.FileName, "attach.txt", "Second nested should have correct filename")
	assert.Contains(t, string(p.Content), "An inline text attachment", "Second nested contains wrong content")
	assert.NotNil(t, p.NextSibling, "Second nested should have a sibling")

	// Third nested
	p = p.NextSibling
	assert.Equal(t, p.Type, "text/plain", "Third nested should have been text")
	assert.Equal(t, p.Disposition, "inline", "Third nested should be inline disposition")
	assert.Equal(t, p.FileName, "attach2.txt", "Third nested should have correct filename")
	assert.Contains(t, string(p.Content), "Another inline text attachment",
		"Third nested contains wrong content")
	assert.Nil(t, p.NextSibling, "Third nested should not have a sibling")
}

func TestMultiBase64Parts(t *testing.T) {
	r := openPart("multibase64.raw")
	p, err := ParseMIME(r)

	// Examine root
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, p.Type, "multipart/mixed", "Expected type to be set")
	assert.Equal(t, len(p.Content), 0, "Root should not have Content")
	assert.NotNil(t, p.FirstChild, "Root should have a FirstChild")
	assert.Nil(t, p.NextSibling, "Root should never have a sibling")

	// Examine first child
	p = p.FirstChild
	assert.Equal(t, p.Type, "text/plain", "First child should have been text")
	assert.Contains(t, string(p.Content), "A text section", "First child contains wrong content")
	assert.NotNil(t, p.NextSibling, "First child should have a sibling")

	// Examine sibling
	p = p.NextSibling
	assert.Equal(t, p.Type, "text/html", "Second child should be html")
	assert.Nil(t, p.NextSibling, "Second child should not have a sibling")
	assert.Nil(t, p.FirstChild, "Second child should not have a child")
	assert.Contains(t, string(p.Content), "<html>",
		"Second child should have <html> as decoded content")
}

// openPart is a test utility function to open a part as a reader
func openPart(filename string) *bufio.Reader {
	// Open test part for parsing
	raw, err := os.Open(filepath.Join("test-data", "parts", filename))
	if err != nil {
		panic(fmt.Sprintf("Failed to open test data: %v", err))
	}

	// Wrap in a buffer
	return bufio.NewReader(raw)
}
