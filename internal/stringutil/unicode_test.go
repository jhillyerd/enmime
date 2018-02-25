package stringutil_test

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/jhillyerd/enmime/internal/stringutil"
)

func TestToASCII(t *testing.T) {
	testCases := []struct {
		input, want string
	}{
		{"", ""},
		{"Yoùr Śtring", "Your String"},
		{"šđčćžŠĐČĆŽŁĲſ", "sdcczSDCCZL__"},
		{"Ötzi's Nationalität èàì", "Otzi's Nationalitat eai"},
	}
	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			got := stringutil.ToASCII(tc.input)
			if got != tc.want {
				t.Errorf("Got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestToASCIILatinExt(t *testing.T) {
	b := make([]byte, 3)
	for r := rune(0x100); r <= 0x17f; r++ {
		n := utf8.EncodeRune(b, r)
		in := string(b[:n])
		out := stringutil.ToASCII(in)
		fmt.Printf("%c %q %q\n", r, in, out)
		if out == "" {
			t.Errorf("ToASCII(%q) returned empty string", in)
		}
		got, _ := utf8.DecodeRuneInString(out)
		if got == utf8.RuneError {
			t.Errorf("ToASCII(%q) returned undecodable rune: %q", in, out)
		}
		if got < 0x21 || 0x7e < got {
			t.Errorf("ToASCII(%q) returned non-ASCII rune: %c (%U)", in, got, got)
		}
	}
}
