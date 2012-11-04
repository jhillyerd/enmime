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
	p, err := parseRoot(r)

	assert.Nil(t, err, "Parsing should not have generated an error")
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, p.Header.Get("Content-Transfer-Encoding"), "7bit",
		"Exepcted Header to have data")
	assert.Equal(t, p.Type, "text/plain", "Expected type to be set")
	assert.Contains(t, string(p.Content), "Test of text/plain section",
		"Expected correct data in p.Content")
}

func TestMultiAlternPart(t *testing.T) {
	r := openPart("multialtern.raw")
	p, err := parseRoot(r)

	// Examine root
	assert.Nil(t, err, "Parsing should not have generated an error")
	assert.NotNil(t, p, "Root node should not be nil")
	assert.Equal(t, p.Type, "multipart/alternative", "Expected type to be set")
	assert.Equal(t, len(p.Content), 0, "Root should not have Content")
	assert.NotNil(t, p.FirstChild, "Root should have a FirstChild")

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
