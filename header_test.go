package enmime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Put the quoted-printable decoder through its paces
func TestQuotedPrintableDecoder(t *testing.T) {
	// Test table data
	var testTable = []struct {
		input, expect string
	}{
		{"", ""},
		{"Hello", "Hello"},
		{"Hello_World", "Hello World"},
		{"Hello=20World", "Hello World"},
		{"Hello=3f", "Hello?"},
		{"Hello=3F", "Hello?"},
	}

	for _, tt := range testTable {
		result, err := decodeQuotedPrintable([]byte(tt.input))
		assert.Nil(t, err)
		assert.Equal(t, tt.expect, string(result),
			"Expected %q, got %q for input %q", tt.expect, result, tt.input)
	}

	// Check malformed quoted chars
	input := []byte("Hello=")
	_, err := decodeQuotedPrintable(input)
	assert.NotNil(t, err)

	input = []byte("Hello=3")
	_, err = decodeQuotedPrintable(input)
	assert.NotNil(t, err)
}

// Check the base64 decoder
func TestBase64Decoder(t *testing.T) {
	input := []byte("SGVsbG8gV29ybGQ=")
	expect := []byte("Hello World")
	result, err := decodeBase64(input)
	assert.Nil(t, err)
	assert.Equal(t, string(expect), string(result))
}

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

// Test control character detection & abort
func TestEncTextControlDetect(t *testing.T) {
	input := "=?US-ASCII?Q?Keith\tMoore?="
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
	// This is not legal, so we expect it to fail
	input := "ab=?US-ASCII?Q?Keith_Moore?=CD"
	expect := input
	result := DecodeHeader(input)
	assert.Equal(t, expect, result)

	// Abutting a MIME header comment is legal
	input = "(=?US-ASCII?Q?Keith_Moore?=)"
	expect = "(Keith Moore)"
	result = DecodeHeader(input)
	assert.Equal(t, expect, result)
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
