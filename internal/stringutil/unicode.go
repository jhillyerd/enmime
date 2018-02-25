package stringutil

import (
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// mapLatinSpecial attempts to map non-accented latin extended runes to ASCII
func mapLatinSpecial(r rune) rune {
	switch r {
	case 'Đ':
		return 'D'
	case 'đ':
		return 'd'
	case 'Ħ':
		return 'H'
	case 'ħ':
		return 'h'
	case 'ĸ':
		return 'K'
	case 'Ŀ':
		return 'L'
	case 'Ł':
		return 'L'
	case 'ŉ':
		return 'n'
	case 'Ŋ':
		return 'N'
	case 'ŋ':
		return 'n'
	case 'Ŧ':
		return 'T'
	case 'ŧ':
		return 't'
	}
	if r > 0x7e {
		return '_'
	}
	return r
}

func ToASCII(s string) string {
	// unicode.Mn: nonspacing marks
	tr := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), runes.Map(mapLatinSpecial),
		norm.NFC)
	r, _, _ := transform.String(tr, s)
	return r
}
