package enmime

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Put the quoted-printable decoder through its paces
func TestQuotedPrintableDecoder(t *testing.T) {
	// Test table data
	var qpTests = []struct {
		input, expect string
	}{
		{"", ""},
		{"Hello", "Hello"},
		{"Hello_World", "Hello World"},
		{"Hello=20World", "Hello World"},
		{"Hello=3f", "Hello?"},
		{"Hello=3F", "Hello?"},
	}

	for _, qp := range qpTests {
		result, err := decodeQuotedPrintable([]byte(qp.input))
		assert.Nil(t, err)
		assert.Equal(t, qp.expect, string(result),
			"got '%s', expected '%v' for '%v'", result, qp.expect, qp.input)
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
	result, err := decodeHeader(input)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Ensure that a string of plain text tokens do not get mangled
func TestPlainTokens(t *testing.T) {
	input := "Testing One two 3 4"
	expect := input
	result, err := decodeHeader(input)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Test control character detection & abort
func TestCharsetControlDetect(t *testing.T) {
	input := "=?US\nASCII?Q?Keith_Moore?="
	expect := input
	result, err := decodeHeader(input)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Test control character detection & abort
func TestEncodingControlDetect(t *testing.T) {
	input := "=?US-ASCII?\r?Keith_Moore?="
	expect := input
	result, err := decodeHeader(input)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Test control character detection & abort
func TestEncTextControlDetect(t *testing.T) {
	input := "=?US-ASCII?Q?Keith\tMoore?="
	expect := input
	result, err := decodeHeader(input)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Test mangled termination
func TestInvalidTermination(t *testing.T) {
	input := "=?US-ASCII?Q?Keith_Moore?!"
	expect := input
	result, err := decodeHeader(input)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Try decoding a simple ASCII quoted-printable encoded word
func TestAsciiQ(t *testing.T) {
	input := "=?US-ASCII?Q?Keith_Moore?="
	expect := "Keith Moore"
	result, err := decodeHeader(input)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Try decoding a simple ASCII quoted-printable encoded word
func TestAsciiB64(t *testing.T) {
	input := "=?US-ASCII?B?SGVsbG8gV29ybGQ=?="
	expect := "Hello World"
	result, err := decodeHeader(input)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Try decoding an embedded ASCII quoted-printable encoded word
func TestEmbeddedAsciiQ(t *testing.T) {
	// This is not legal, so we expect it to fail
	input := "ab=?US-ASCII?Q?Keith_Moore?=CD"
	expect := input
	result, err := decodeHeader(input)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)

	// Abutting a MIME header comment is legal
	input = "(=?US-ASCII?Q?Keith_Moore?=)"
	expect = "(Keith Moore)"
	result, err = decodeHeader(input)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Spacing rules from RFC 2047
func TestSpacing(t *testing.T) {
	var spTable = []struct {
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

	for _, sp := range spTable {
		result, err := decodeHeader(sp.input)
		assert.Nil(t, err)
		assert.Equal(t, sp.expect, result,
			"got '%v', expected '%v' for '%v'", result, sp.expect, sp.input)
	}
}
