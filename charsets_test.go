package enmime

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test an invalid character set with the CharsetReader
func TestInvalidCharsetReader(t *testing.T) {
	inputReader := strings.NewReader("unused")
	outputReader, err := NewCharsetReader("INVALIDcharsetZZZ", inputReader)
	assert.Nil(t, outputReader, "outputReader should be nil")
	assert.NotNil(t, err, "err should not be nil")
}

// Test some different character sets with the CharsetReader
func TestCharsetReader(t *testing.T) {
	var testTable = []struct {
		charset string
		input   []byte
		expect  string
	}{
		{"utf-8", []byte("abcABC\u2014"), "abcABC\u2014"},
		{"windows-1250", []byte{'a', 'Z', 0x96}, "aZ\u2013"},
		{"big5", []byte{0xa1, 0x5d, 0xa1, 0x61, 0xa1, 0x71}, "\uff08\uff5b\u3008"},
	}

	for _, tt := range testTable {
		inputReader := bytes.NewReader(tt.input)
		outputReader, err := NewCharsetReader(tt.charset, inputReader)
		assert.Nil(t, err)
		result, err := ioutil.ReadAll(outputReader)
		assert.Nil(t, err)
		assert.Equal(t, tt.expect, string(result),
			"Expected %q, got %q for input %q", tt.expect, result, tt.input)
	}
}
