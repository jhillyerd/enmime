package enmime

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainTextPart(t *testing.T) {
	r := openPart("textplain.raw")
	p, err := ParseMIME(r)

	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, "7bit", p.Header().Get("Content-Transfer-Encoding"),
		"Exepcted Header to have data")
	assert.Equal(t, "text/plain", p.ContentType(), "Expected type to be set")
	assert.Contains(t, string(p.Content()), "Test of text/plain section",
		"Expected correct data in p.Content")
	assert.Nil(t, p.NextSibling(), "Root should never have a sibling")
}

func TestQuotedPrintablePart(t *testing.T) {
	r := openPart("quoted-printable.raw")
	p, err := ParseMIME(r)

	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, "quoted-printable", p.Header().Get("Content-Transfer-Encoding"),
		"Exepcted Header to have data")
	assert.Equal(t, "text/plain", p.ContentType(), "Expected type to be set")
	assert.Equal(t, "Start=ABC=Finish", string(p.Content()),
		"Expected correct data in p.Content")
	assert.Nil(t, p.NextSibling(), "Root should never have a sibling")
}

func TestMultiAlternParts(t *testing.T) {
	r := openPart("multialtern.raw")
	p, err := ParseMIME(r)

	// Examine root
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, "multipart/alternative", p.ContentType(), "Expected type to be set")
	assert.Equal(t, 0, len(p.Content()), "Root should not have Content")
	assert.NotNil(t, p.FirstChild(), "Root should have a FirstChild")
	assert.Nil(t, p.NextSibling(), "Root should never have a sibling")

	// Examine first child
	p = p.FirstChild()
	assert.Equal(t, "text/plain", p.ContentType(), "First child should have been text")
	assert.Contains(t, string(p.Content()), "A text section", "First child contains wrong content")
	assert.NotNil(t, p.NextSibling(), "First child should have a sibling")

	// Examine sibling
	p = p.NextSibling()
	assert.Equal(t, "text/html", p.ContentType(), "Second child should have been html")
	assert.Contains(t, string(p.Content()), "An HTML section", "Second child contains wrong content")
	assert.Nil(t, p.NextSibling(), "Second child should not have a sibling")
}

func TestMultiMixedParts(t *testing.T) {
	r := openPart("multimixed.raw")
	p, err := ParseMIME(r)

	// Examine root
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, "multipart/mixed", p.ContentType(), "Expected type to be set")
	assert.Equal(t, 0, len(p.Content()), "Root should not have Content")
	assert.NotNil(t, p.FirstChild(), "Root should have a FirstChild")
	assert.Nil(t, p.NextSibling(), "Root should never have a sibling")

	// Examine first child
	p = p.FirstChild()
	assert.Equal(t, "text/plain", p.ContentType(), "First child should have been text")
	assert.Contains(t, string(p.Content()), "Section one", "First child contains wrong content")
	assert.NotNil(t, p.NextSibling(), "First child should have a sibling")

	// Examine sibling
	p = p.NextSibling()
	assert.Equal(t, "text/plain", p.ContentType(), "Second child should have been html")
	assert.Contains(t, string(p.Content()), "Section two", "Second child contains wrong content")
	assert.Nil(t, p.NextSibling(), "Second child should not have a sibling")
}

func TestMultiOtherParts(t *testing.T) {
	r := openPart("multiother.raw")
	p, err := ParseMIME(r)

	// Examine root
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, "multipart/x-enmime", p.ContentType(), "Expected type to be set")
	assert.Equal(t, 0, len(p.Content()), "Root should not have Content")
	assert.NotNil(t, p.FirstChild(), "Root should have a FirstChild")
	assert.Nil(t, p.NextSibling(), "Root should never have a sibling")

	// Examine first child
	p = p.FirstChild()
	assert.Equal(t, "text/plain", p.ContentType(), "First child should have been text")
	assert.Contains(t, string(p.Content()), "Section one", "First child contains wrong content")
	assert.NotNil(t, p.NextSibling(), "First child should have a sibling")

	// Examine sibling
	p = p.NextSibling()
	assert.Equal(t, "text/plain", p.ContentType(), "Second child should have been html")
	assert.Contains(t, string(p.Content()), "Section two", "Second child contains wrong content")
	assert.Nil(t, p.NextSibling(), "Second child should not have a sibling")
}

func TestNestedAlternParts(t *testing.T) {
	r := openPart("nestedmulti.raw")
	p, err := ParseMIME(r)

	// Examine root
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, "multipart/alternative", p.ContentType(), "Expected type to be set")
	assert.Equal(t, 0, len(p.Content()), "Root should not have Content")
	assert.NotNil(t, p.FirstChild(), "Root should have a FirstChild")
	assert.Nil(t, p.NextSibling(), "Root should never have a sibling")

	// Examine first child
	p = p.FirstChild()
	assert.Equal(t, "text/plain", p.ContentType(), "First child should have been text")
	assert.Contains(t, string(p.Content()), "A text section", "First child contains wrong content")
	assert.NotNil(t, p.NextSibling(), "First child should have a sibling")

	// Examine sibling
	p = p.NextSibling()
	assert.Equal(t, "multipart/related", p.ContentType(), "Second child should have been another multipart")
	assert.Equal(t, 0, len(p.Content()), "Second child should not have Content")
	assert.Nil(t, p.NextSibling(), "Second child should not have a sibling")
	assert.NotNil(t, p.FirstChild(), "Second child should have a child")

	// First nested
	p = p.FirstChild()
	assert.Equal(t, "text/html", p.ContentType(), "First nested should have been html")
	assert.Contains(t, string(p.Content()), "An HTML section", "First nested contains wrong content")
	assert.NotNil(t, p.NextSibling(), "First nested should have a sibling")

	// Second nested
	p = p.NextSibling()
	assert.Equal(t, "text/plain", p.ContentType(), "Second nested should have been text")
	assert.Equal(t, "inline", p.Disposition(), "Second nested should be inline disposition")
	assert.Equal(t, "attach.txt", p.FileName(), "Second nested should have correct filename")
	assert.Contains(t, string(p.Content()), "An inline text attachment", "Second nested contains wrong content")
	assert.NotNil(t, p.NextSibling(), "Second nested should have a sibling")

	// Third nested
	p = p.NextSibling()
	assert.Equal(t, "text/plain", p.ContentType(), "Third nested should have been text")
	assert.Equal(t, "inline", p.Disposition(), "Third nested should be inline disposition")
	assert.Equal(t, "attach2.txt", p.FileName(), "Third nested should have correct filename")
	assert.Contains(t, string(p.Content()), "Another inline text attachment",
		"Third nested contains wrong content")
	assert.Nil(t, p.NextSibling(), "Third nested should not have a sibling")
}

func TestMultiBase64Parts(t *testing.T) {
	r := openPart("multibase64.raw")
	p, err := ParseMIME(r)

	// Examine root
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, "multipart/mixed", p.ContentType(), "Expected type to be set")
	assert.Equal(t, 0, len(p.Content()), "Root should not have Content")
	assert.NotNil(t, p.FirstChild(), "Root should have a FirstChild")
	assert.Nil(t, p.NextSibling(), "Root should never have a sibling")

	// Examine first child
	p = p.FirstChild()
	assert.Equal(t, "text/plain", p.ContentType(), "First child should have been text")
	assert.Contains(t, string(p.Content()), "A text section", "First child contains wrong content")
	assert.NotNil(t, p.NextSibling(), "First child should have a sibling")

	// Examine sibling
	p = p.NextSibling()
	assert.Equal(t, "text/html", p.ContentType(), "Second child should be html")
	assert.Nil(t, p.NextSibling(), "Second child should not have a sibling")
	assert.Nil(t, p.FirstChild(), "Second child should not have a child")
	assert.Contains(t, string(p.Content()), "<html>",
		"Second child should have <html> as decoded content")
}

func TestBadBoundaryTerm(t *testing.T) {
	r := openPart("badboundary.raw")
	p, err := ParseMIME(r)

	// Examine root
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, "multipart/alternative", p.ContentType(), "Expected type to be set")

	// Examine first child
	p = p.FirstChild()
	assert.Equal(t, "text/plain", p.ContentType(), "First child should have been text")
	assert.NotNil(t, p.NextSibling(), "First child should have a sibling")

	// Examine sibling
	p = p.NextSibling()
	assert.Equal(t, "text/html", p.ContentType(), "Second child should have been html")
	assert.Contains(t, string(p.Content()), "An HTML section", "Second child contains wrong content")
	assert.Nil(t, p.NextSibling(), "Second child should not have a sibling")
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
