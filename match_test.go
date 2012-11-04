package enmime

import (
	"github.com/stretchrcom/testify/assert"
	"testing"
)

func TestBreadthMatchFirst(t *testing.T) {
	r := openPart("nestedmulti.raw")
	root, err := ParseMIME(r)

	// Sanity check
	if !assert.Nil(t, err, "Parsing should not have generated an error") {
		t.FailNow()
	}
	assert.NotNil(t, root, "Root node should not be nil")

	p := root.BreadthMatchFirst(func(pt *MIMEPart) bool { return pt.Type == "text/plain" })
	assert.NotNil(t, p, "BreathMatchFirst should have returned a result for text/plain")
	assert.Contains(t, string(p.Content), "A text section")

	p = root.BreadthMatchFirst(func(pt *MIMEPart) bool { return pt.Type == "text/html" })
	assert.NotNil(t, p, "BreathMatchFirst should have returned a result for text/html")
	assert.Contains(t, string(p.Content), "An HTML section")
}

