package enmime

import (
	"testing"
)

// Ensure that a single plain text token passes unharmed
func TestPlainSingleToken(t *testing.T) {
	in := "Test"
	want := in
	got := DecodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Ensure that a string of plain text tokens do not get mangled
func TestPlainTokens(t *testing.T) {
	in := "Testing One two 3 4"
	want := in
	got := DecodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Test control character detection & abort
func TestCharsetControlDetect(t *testing.T) {
	in := "=?US\nASCII?Q?Keith_Moore?="
	want := in
	got := DecodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Test control character detection & abort
func TestEncodingControlDetect(t *testing.T) {
	in := "=?US-ASCII?\r?Keith_Moore?="
	want := in
	got := DecodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Test mangled termination
func TestInvalidTermination(t *testing.T) {
	in := "=?US-ASCII?Q?Keith_Moore?!"
	want := in
	got := DecodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Try decoding a simple ASCII quoted-printable encoded word
func TestAsciiQ(t *testing.T) {
	in := "=?US-ASCII?Q?Keith_Moore?="
	want := "Keith Moore"
	got := DecodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Try decoding a simple ASCII quoted-printable encoded word
func TestAsciiB64(t *testing.T) {
	in := "=?US-ASCII?B?SGVsbG8gV29ybGQ=?="
	want := "Hello World"
	got := DecodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Try decoding an embedded ASCII quoted-printable encoded word
func TestEmbeddedAsciiQ(t *testing.T) {
	var testTable = []struct {
		in, want string
	}{
		// Abutting a MIME header comment is legal
		{"(=?US-ASCII?Q?Keith_Moore?=)", "(Keith Moore)"},
		// The entire header does not need to be encoded
		{"(Keith =?US-ASCII?Q?Moore?=)", "(Keith Moore)"},
	}

	for _, tt := range testTable {
		got := DecodeHeader(tt.in)
		if got != tt.want {
			t.Errorf("DecodeHeader(%q) == %q, want: %q", tt.in, got, tt.want)
		}
	}
}

// Spacing rules from RFC 2047
func TestSpacing(t *testing.T) {
	var testTable = []struct {
		in, want string
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
		got := DecodeHeader(tt.in)
		if got != tt.want {
			t.Errorf("DecodeHeader(%q) == %q, want: %q", tt.in, got, tt.want)
		}
	}
}

// Test some different character sets
func TestCharsets(t *testing.T) {
	var testTable = []struct {
		in, want string
	}{
		{"=?utf-8?q?abcABC_=24_=c2=a2_=e2=82=ac?=", "abcABC $ \u00a2 \u20ac"},
		{"=?iso-8859-1?q?#=a3_c=a9_r=ae_u=b5?=", "#\u00a3 c\u00a9 r\u00ae u\u00b5"},
		{"=?big5?q?=a1=5d_=a1=61_=a1=71?=", "\uff08 \uff5b \u3008"},
	}

	for _, tt := range testTable {
		got := DecodeHeader(tt.in)
		if got != tt.want {
			t.Errorf("DecodeHeader(%q) == %q, want: %q", tt.in, got, tt.want)
		}
	}
}

// Test re-encoding to base64
func TestDecodeToUTF8Base64Header(t *testing.T) {
	var testTable = []struct {
		in, want string
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
		got := DecodeToUTF8Base64Header(tt.in)
		if got != tt.want {
			t.Errorf("DecodeHeader(%q) == %q, want: %q", tt.in, got, tt.want)
		}
	}
}
