package coding_test

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime/internal/coding"
)

// Test an invalid character set with the CharsetReader
func TestInvalidCharsetReader(t *testing.T) {
	inputReader := strings.NewReader("unused")
	outputReader, err := coding.NewCharsetReader("INVALIDcharsetZZZ", inputReader)
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
		outputReader, err := coding.NewCharsetReader(tt.charset, inputReader)
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

// Search for character set info inside of HTML
func TestFindCharsetInHTML(t *testing.T) {
	var ttable = []struct {
		input, want string
	}{
		{`<meta charset="UTF-8">`, "UTF-8"},
		{`<meta attrib="value" charset="us-ascii"/>`, "us-ascii"},
		{`<meta charset=big5 other=value>`, "big5"},
		{`<meta charset=us-ascii>`, "us-ascii"},
		{`<meta charset=windows-1250/>`, "windows-1250"},
	}

	for _, tt := range ttable {
		got := coding.FindCharsetInHTML(tt.input)
		if got != tt.want {
			t.Errorf("Got: %q, want: %q, for: %q", got, tt.want, tt.input)
		}
	}
}
