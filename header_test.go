package enmime

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Put the quoted-printable decoder through its paces
func TestQuotedPrintableDecoder(t *testing.T) {
	var input, expect, result []byte
	var err error

	input = []byte("")
	expect = []byte("")
	result, err = decodeQuotedPrintable(input)
	assert.Nil(t, err)
	assert.Equal(t, string(expect), string(result))

	input = []byte("Hello")
	expect = []byte("Hello")
	result, err = decodeQuotedPrintable(input)
	assert.Nil(t, err)
	assert.Equal(t, string(expect), string(result))

	input = []byte("Hello_World")
	expect = []byte("Hello World")
	result, err = decodeQuotedPrintable(input)
	assert.Nil(t, err)
	assert.Equal(t, string(expect), string(result))

	input = []byte("Hello=20World")
	expect = []byte("Hello World")
	result, err = decodeQuotedPrintable(input)
	assert.Nil(t, err)
	assert.Equal(t, string(expect), string(result))

	input = []byte("Hello=3f")
	expect = []byte("Hello?")
	result, err = decodeQuotedPrintable(input)
	assert.Nil(t, err)
	assert.Equal(t, string(expect), string(result))

	input = []byte("Hello=3F")
	expect = []byte("Hello?")
	result, err = decodeQuotedPrintable(input)
	assert.Nil(t, err)
	assert.Equal(t, string(expect), string(result))

	input = []byte("Hello=")
	result, err = decodeQuotedPrintable(input)
	assert.NotNil(t, err)

	input = []byte("Hello=3")
	result, err = decodeQuotedPrintable(input)
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
	input := "ab=?US-ASCII?Q?Keith_Moore?=CD"
	expect := "abKeith MooreCD"
	result, err := decodeHeader(input)

	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

// Spacing rules from RFC 2047
func TestSpacing(t *testing.T) {
	var input, expect, result string
	var err error

	input = "(=?ISO-8859-1?Q?a?=)"
	expect = "(a)"
	result, err = decodeHeader(input)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)

	input = "(=?ISO-8859-1?Q?a?= b)"
	expect = "(a b)"
	result, err = decodeHeader(input)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)

	input = "(=?ISO-8859-1?Q?a?= =?ISO-8859-1?Q?b?=)"
	expect = "(ab)"
	result, err = decodeHeader(input)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)

	input = "(=?ISO-8859-1?Q?a?=  =?ISO-8859-1?Q?b?=)"
	expect = "(ab)"
	result, err = decodeHeader(input)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)

	input = "(=?ISO-8859-1?Q?a?=\r\n  =?ISO-8859-1?Q?b?=)"
	expect = "(ab)"
	result, err = decodeHeader(input)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)

	input = "(=?ISO-8859-1?Q?a_b?=)"
	expect = "(a b)"
	result, err = decodeHeader(input)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)

	input = "(=?ISO-8859-1?Q?a?= =?ISO-8859-2?Q?_b?=)"
	expect = "(a b)"
	result, err = decodeHeader(input)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

