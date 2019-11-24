package stringutil

import (
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var latinSpecialMap = map[rune]rune{
	'Đ': 'D',
	'đ': 'd',
	'Ħ': 'H',
	'ħ': 'h',
	'ĸ': 'K',
	'Ŀ': 'L',
	'Ł': 'L',
	'ŉ': 'n',
	'Ŋ': 'N',
	'ŋ': 'n',
	'Ŧ': 'T',
	'ŧ': 't',
}

// mapLatinSpecial attempts to map non-accented latin extended runes to ASCII
func mapLatinSpecial(r rune) rune {
	if v, ok := latinSpecialMap[r]; ok {
		return v
	}
	if r > 0x7e {
		return '_'
	}
	return r
}

// ToASCII converts unicode to ASCII by stripping accents and converting some special characters
// into their ASCII approximations.  Anything else will be replaced with an underscore.
func ToASCII(s string) string {
	// unicode.Mn: nonspacing marks
	tr := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), runes.Map(mapLatinSpecial),
		norm.NFC)
	r, _, _ := transform.String(tr, s)
	return r
}
