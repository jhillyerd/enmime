package stringutil

import (
	"bytes"
	"net/mail"
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
