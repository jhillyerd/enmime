package stringutil

const escape = '\\'

// SplitQuoted splits a string, ignoring separators present inside of quoted runs.  Separators
// cannot be escaped outside of quoted runs, the escaping will be ignored.
//
// Quotes are preserved in result, but the separators are removed.
func SplitQuoted(s string, sep rune, quote rune) []string {
	a := make([]string, 0, 8)
	quoted := false
	escaped := false
	p := 0
	for i, c := range s {
		if c == escape {
			// Escape can escape itself.
			escaped = !escaped
			continue
		}
		if c == quote {
			quoted = !quoted
			continue
		}
		escaped = false
		if !quoted && c == sep {
			a = append(a, s[p:i])
			p = i + 1
		}
	}

	if quoted && quote != 0 {
		// s contained an unterminated quoted-run, re-split without quoting.
		return SplitQuoted(s, sep, rune(0))
	}

	return append(a, s[p:])
}
