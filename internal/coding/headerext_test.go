package coding

import (
	"testing"
)

func TestDecodePlainPassthrough(t *testing.T) {
	var ttable = []string{
		"Test",
		"Testing One two 3 4",
	}

	for _, in := range ttable {
		t.Run(in, func(t *testing.T) {
			got := DecodeExtHeader(in)
			if got != in {
				t.Errorf("DecodeExtHeader(%q) == %q, want: %q", in, got, in)
			}
		})
	}
}

func TestDecodeFailurePassthrough(t *testing.T) {
	var ttable = []struct {
		label, in string
	}{
		{
			label: "Newline detection & abort",
			in:    "=?US\nASCII?Q?Keith_Moore?=",
		},
		{
			label: "Carriage return detection & abort",
			in:    "=?US-ASCII?\r?Keith_Moore?=",
		},
		{
			label: "Invalid termination",
			in:    "=?US-ASCII?Q?Keith_Moore?!",
		},
	}

	for _, tt := range ttable {
		t.Run(tt.in, func(t *testing.T) {
			got := DecodeExtHeader(tt.in)
			if got != tt.in {
				t.Errorf("DecodeExtHeader(%q) == %q, want: %q", tt.in, got, tt.in)
			}
		})
	}
}

func TestDecodeAsciiB64(t *testing.T) {
	var ttable = []struct {
		in, want string
	}{
		// Simple ASCII quoted-printable encoded word
		{"=?US-ASCII?B?SGVsbG8gV29ybGQ=?=", "Hello World"},
		// Abutting a MIME header comment is legal
		{"(=?US-ASCII?B?SGVsbG8gV29ybGQ=?=)", "(Hello World)"},
		// The entire header does not need to be encoded
		{"(Prefix =?US-ASCII?B?SGVsbG8gV29ybGQ=?=)", "(Prefix Hello World)"},
	}

	for _, tt := range ttable {
		t.Run(tt.in, func(t *testing.T) {
			got := DecodeExtHeader(tt.in)
			if got != tt.want {
				t.Errorf("DecodeExtHeader(%q) == %q, want: %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDecodeAsciiQ(t *testing.T) {
	var ttable = []struct {
		in, want string
	}{
		// Simple ASCII QP encoded word
		{"=?US-ASCII?Q?Keith_Moore?=", "Keith Moore"},
		// Abutting a MIME header comment is legal
		{"(=?US-ASCII?Q?Keith_Moore?=)", "(Keith Moore)"},
		// The entire header does not need to be encoded
		{"(Keith =?US-ASCII?Q?Moore?=)", "(Keith Moore)"},
	}

	for _, tt := range ttable {
		t.Run(tt.in, func(t *testing.T) {
			got := DecodeExtHeader(tt.in)
			if got != tt.want {
				t.Errorf("DecodeExtHeader(%q) == %q, want: %q", tt.in, got, tt.want)
			}
		})
	}
}

// Spacing rules from RFC 2047
func TestDecodeSpacing(t *testing.T) {
	var ttable = []struct {
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

	for _, tt := range ttable {
		t.Run(tt.in, func(t *testing.T) {
			got := DecodeExtHeader(tt.in)
			if got != tt.want {
				t.Errorf("DecodeExtHeader(%q) == %q, want: %q", tt.in, got, tt.want)
			}
		})
	}
}

// Test some different character sets
func TestDecodeCharsets(t *testing.T) {
	var ttable = []struct {
		in, want string
	}{
		{"=?utf-8?q?abcABC_=24_=c2=a2_=e2=82=ac?=", "abcABC $ \u00a2 \u20ac"},
		{"=?iso-8859-1?q?#=a3_c=a9_r=ae_u=b5?=", "#\u00a3 c\u00a9 r\u00ae u\u00b5"},
		{"=?big5?q?=a1=5d_=a1=61_=a1=71?=", "\uff08 \uff5b \u3008"},
	}

	for _, tt := range ttable {
		t.Run(tt.in, func(t *testing.T) {
			got := DecodeExtHeader(tt.in)
			if got != tt.want {
				t.Errorf("DecodeExtHeader(%q) == %q, want: %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRfc2047Decode(t *testing.T) {
	var ttable = []struct {
		label, in, want string
	}{
		{"pass through", "plain text", "plain text"},
		{"quoted", "=?US-ASCII?q?Hello=20World?=", "Hello World"},
		{"base64", "=?US-ASCII?b?SGVsbG8gV29ybGQ=?=", "Hello World"},
		{"nested qp+b64", "=?utf-8?b?PT9VUy1BU0NJST9xP0hlbGxvPTIwV29ybGQ/PQ==?=", "Hello World"},
	}

	for _, tt := range ttable {
		t.Run(tt.label, func(t *testing.T) {
			got := RFC2047Decode(tt.in)
			if got != tt.want {
				t.Errorf("RFC2047Decode(%q) == %q, want: %q", tt.in, got, tt.want)
			}
		})
	}
}
