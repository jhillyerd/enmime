package stringutil

const escape = '\\'

// SplitQuoted slices s into all substrings separated by sep and returns a slice of
// the substrings between those separators.
//
// If s does not contain sep and sep is not empty, SplitQuoted returns a
// slice of length 1 whose only element is s.
//
// It ignores sep present inside quoted runs.
func SplitQuoted(s string, sep rune, quote rune) []string {
	return splitQuoted(s, sep, quote, false)
}

// SplitAfterQuoted slices s into all substrings after each instance of sep and
// returns a slice of those substrings.
//
// If s does not contain sep and sep is not empty, SplitAfterQuoted returns
// a slice of length 1 whose only element is s.
//
// It ignores sep present inside quoted runs.
func SplitAfterQuoted(s string, sep rune, quote rune) []string {
	return splitQuoted(s, sep, quote, true)
}

func splitQuoted(s string, sep rune, quote rune, preserveSep bool) []string {
	ixs := FindQuoted(s, sep, quote)
	if len(ixs) == 0 {
		return []string{s}
	}

	start := 0
	result := make([]string, 0, len(ixs)+1)
	for _, ix := range ixs {
		end := ix
		if preserveSep {
			end++
		}
		result = append(result, s[start:end])
		start = ix + 1
	}

	return append(result, s[start:])
}
