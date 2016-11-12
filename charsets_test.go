package enmime

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

// Test an invalid character set with the CharsetReader
func TestInvalidCharsetReader(t *testing.T) {
	inputReader := strings.NewReader("unused")
	outputReader, err := NewCharsetReader("INVALIDcharsetZZZ", inputReader)
	if outputReader != nil {
		t.Error("outputReader should be nil")
	}
	if err == nil {
		t.Error("err should not be nil")
	}
}

// Test some different character sets with the CharsetReader
func TestCharsetReader(t *testing.T) {
	var testTable = []struct {
		charset string
		input   []byte
		want    string
	}{
		{"utf-8", []byte("abcABC\u2014"), "abcABC\u2014"},
		{"windows-1250", []byte{'a', 'Z', 0x96}, "aZ\u2013"},
		{"big5", []byte{0xa1, 0x5d, 0xa1, 0x61, 0xa1, 0x71}, "\uff08\uff5b\u3008"},
	}

	for _, tt := range testTable {
		inputReader := bytes.NewReader(tt.input)
		outputReader, err := NewCharsetReader(tt.charset, inputReader)
		if err != nil {
			t.Error("err should be nil, got:", err)
		}
		result, err := ioutil.ReadAll(outputReader)
		if err != nil {
			t.Error("err should be nil, got:", err)
		}
		got := string(result)
		if got != tt.want {
			t.Errorf("NewCharsetReader(%q, %q) = %q, want: %q", tt.charset, tt.input, got, tt.want)
		}
	}
}
