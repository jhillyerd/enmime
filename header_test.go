package enmime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensure that a single plain text token passes unharmed
func TestPlainSingleToken(t *testing.T) {
	input := "Test"
	expect := input
	result := DecodeHeader(input)
	assert.Equal(t, expect, result)
}

// Ensure that a string of plain text tokens do not get mangled
func TestPlainTokens(t *testing.T) {
	input := "Testing One two 3 4"
	expect := input
	result := DecodeHeader(input)
	assert.Equal(t, expect, result)
}

// Test control character detection & abort
func TestCharsetControlDetect(t *testing.T) {
	input := "=?US\nASCII?Q?Keith_Moore?="
	expect := input
	result := DecodeHeader(input)
	assert.Equal(t, expect, result)
}

// Test control character detection & abort
func TestEncodingControlDetect(t *testing.T) {
	input := "=?US-ASCII?\r?Keith_Moore?="
	expect := input
	result := DecodeHeader(input)
	assert.Equal(t, expect, result)
}

// Test mangled termination
func TestInvalidTermination(t *testing.T) {
	input := "=?US-ASCII?Q?Keith_Moore?!"
	expect := input
	result := DecodeHeader(input)
	assert.Equal(t, expect, result)
}

// Try decoding a simple ASCII quoted-printable encoded word
func TestAsciiQ(t *testing.T) {
	input := "=?US-ASCII?Q?Keith_Moore?="
	expect := "Keith Moore"
	result := DecodeHeader(input)
	assert.Equal(t, expect, result)
}

// Try decoding a simple ASCII quoted-printable encoded word
func TestAsciiB64(t *testing.T) {
	input := "=?US-ASCII?B?SGVsbG8gV29ybGQ=?="
	expect := "Hello World"
	result := DecodeHeader(input)
	assert.Equal(t, expect, result)
}

// Try decoding an embedded ASCII quoted-printable encoded word
func TestEmbeddedAsciiQ(t *testing.T) {
	var testTable = []struct {
		input, expect string
	}{
		// Abutting a MIME header comment is legal
		{"(=?US-ASCII?Q?Keith_Moore?=)", "(Keith Moore)"},
		// The entire header does not need to be encoded
		{"(Keith =?US-ASCII?Q?Moore?=)", "(Keith Moore)"},
	}

	for _, tt := range testTable {
		result := DecodeHeader(tt.input)
		assert.Equal(t, tt.expect, result,
			"Expected %q, got %q for input %q", tt.expect, result, tt.input)
	}
}

// Spacing rules from RFC 2047
func TestSpacing(t *testing.T) {
	var testTable = []struct {
		input, expect string
	}{
		{"(=?ISO-8859-1?Q?a?=)", "(a)"},
		{"(=?ISO-8859-1?Q?a?= b)", "(a b)"},
		{"(=?ISO-8859-1?Q?a?= =?ISO-8859-1?Q?b?=)", "(ab)"},
		{"(=?ISO-8859-1?Q?a?=  =?ISO-8859-1?Q?b?=)", "(ab)"},
		{"(=?ISO-8859-1?Q?a?=\r\n  =?ISO-8859-1?Q?b?=)", "(ab)"},
		{"(=?ISO-8859-1?Q?a_b?=)", "(a b)"},
		{"(=?ISO-8859-1?Q?a?= =?ISO-8859-2?Q?_b?=)", "(a b)"},
	}

	for _, tt := range testTable {
		result := DecodeHeader(tt.input)
		assert.Equal(t, tt.expect, result,
			"Expected %q, got %q for input %q", tt.expect, result, tt.input)
	}
}

// Test some different character sets
func TestCharsets(t *testing.T) {
	var testTable = []struct {
		input, expect string
	}{
		{"=?utf-8?q?abcABC_=24_=c2=a2_=e2=82=ac?=", "abcABC $ \u00a2 \u20ac"},
		{"=?iso-8859-1?q?#=a3_c=a9_r=ae_u=b5?=", "#\u00a3 c\u00a9 r\u00ae u\u00b5"},
		{"=?big5?q?=a1=5d_=a1=61_=a1=71?=", "\uff08 \uff5b \u3008"},
	}

	for _, tt := range testTable {
		result := DecodeHeader(tt.input)
		assert.Equal(t, tt.expect, result,
			"Expected %q, got %q for input %q", tt.expect, result, tt.input)
	}
}

// Test re-encoding to base64
func TestDecodeToUTF8Base64Header(t *testing.T) {
	var testTable = []struct {
		input, expect string
	}{
		{"no encoding", "no encoding"},
		{"=?utf-8?q?abcABC_=24_=c2=a2_=e2=82=ac?=", "=?UTF-8?b?YWJjQUJDICQgwqIg4oKs?="},
		{"=?iso-8859-1?q?#=a3_c=a9_r=ae_u=b5?=", "=?UTF-8?b?I8KjIGPCqSBywq4gdcK1?="},
		{"=?big5?q?=a1=5d_=a1=61_=a1=71?=", "=?UTF-8?b?77yIIO+9myDjgIg=?="},
		// Must respect separate tokens
		{"=?UTF-8?Q?Miros=C5=82aw?= <u@h>", "=?UTF-8?b?TWlyb3PFgmF3?= <u@h>"},
		{"First Last <u@h> (=?iso-8859-1?q?#=a3_c=a9_r=ae_u=b5?=)",
			"First Last <u@h> (=?UTF-8?b?I8KjIGPCqSBywq4gdcK1?=)"},
	}

	for _, tt := range testTable {
		result := DecodeToUTF8Base64Header(tt.input)
		assert.Equal(t, tt.expect, result,
			"Expected %q, got %q for input %q", tt.expect, result, tt.input)
	}
}
