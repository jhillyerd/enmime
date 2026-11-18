package stringutil

import (
	"bytes"
	"net/mail"
	"strings"
)

// JoinAddress formats a slice of Address structs such that they can be used in a To or Cc header.
func JoinAddress(addrs []mail.Address) string {
	if len(addrs) == 0 {
		return ""
	}
	buf := &bytes.Buffer{}
	for i, a := range addrs {
		if i > 0 {
			_, _ = buf.WriteString(", ")
		}
		_, _ = buf.WriteString(a.String())
	}
	return buf.String()
}

// EnsureCommaDelimitedAddresses is used by AddressList to ensure that address lists are properly
// delimited.
func EnsureCommaDelimitedAddresses(s string) string {
	// This normalizes the whitespace, but may interfere with CFWS (comments with folding whitespace)
	// RFC-5322 3.4.0:
	//      because some legacy implementations interpret the comment,
	//      comments generally SHOULD NOT be used in address fields
	//      to avoid confusing such implementations.
	s = strings.Join(strings.Fields(s), " ")

	inQuotes := false
	inDomain := false
	escapeSequence := false
	inAngles := false
	inGroup := false
	commentDepth := 0
	sb := strings.Builder{}
	skipTo := 0
	for i, r := range s {
		if i < skipTo {
			continue
		}
		if escapeSequence {
			escapeSequence = false
			sb.WriteRune(r)
			continue
		}
		if commentDepth > 0 {
			// Keep comments as they are, so a colon in "(role:admin)" doesn't open a group.
			switch r {
			case '\\':
				escapeSequence = true
			case '(':
				commentDepth++
			case ')':
				commentDepth--
			}
			sb.WriteRune(r)
			continue
		}
		if r == '(' && !inQuotes && !inAngles {
			commentDepth++
			sb.WriteRune(r)
			continue
		}
		if r == '"' {
			inQuotes = !inQuotes
			sb.WriteRune(r)
			continue
		}

		if r == '<' {
			inAngles = true
			sb.WriteRune(r)
			continue
		}

		if r == '>' {
			inAngles = false
			sb.WriteRune(r)
			continue
		}

		if inQuotes {
			if r == '\\' {
				escapeSequence = true
				sb.WriteRune(r)
				continue
			}
		} else {
			if r == ':' && !inAngles && !inDomain {
				// Start of a group, as in "Team: a@example.com, b@example.com;".
				inGroup = true
				sb.WriteRune(r)
				continue
			}
			if r == ';' && inGroup && !inAngles {
				// Keep the semicolon that ends the group along with any comment after it,
				// and separate the group from an address that follows.
				inGroup = false
				inDomain = false
				sb.WriteRune(r)
				rest := skipCFWS(s[i+1:])
				comment := strings.TrimRight(s[i+1:len(s)-len(rest)], " ")
				sb.WriteString(comment)
				skipTo = i + 1 + len(comment)
				if rest != "" && rest[0] != ',' {
					sb.WriteRune(',')
				}
				continue
			}
			if r == '@' {
				inDomain = true
				sb.WriteRune(r)
				continue
			}
			if inDomain {
				if r == ';' {
					inDomain = false
					if i == len(s)-1 {
						// omit trailing semicolon
						continue
					}

					sb.WriteRune(',')
					continue
				}
				if r == ',' {
					inDomain = false
					sb.WriteRune(r)
					continue
				}
				if r == ' ' && !inAngles {
					inDomain = false
					sb.WriteRune(',')
					sb.WriteRune(r)
					continue
				}
			}

			if inAngles && r == ' ' {
				continue
			}
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// skipCFWS returns s without its leading spaces and comments.
func skipCFWS(s string) string {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\\' && depth > 0:
			i++
		case c == '(':
			depth++
		case c == ')' && depth > 0:
			depth--
		case c != ' ' && depth == 0:
			return s[i:]
		}
	}
	return ""
}
