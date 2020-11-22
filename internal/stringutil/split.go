package stringutil

// SplitQuoted splits a string, ignoring seperators present inside of quoted runs.
func SplitQuoted(s string, sep rune, quote rune) []string {
	a := make([]string, 0, 8)
	quoted := false
	p := 0
	for i, c := range s {
		if c == quote {
			quoted = !quoted
			continue
		}
		if !quoted && c == sep {
			a = append(a, s[p:i])
			p = i + 1
		}
	}
	return append(a, s[p:])
}
