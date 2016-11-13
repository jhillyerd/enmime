package enmime

import (
	"fmt"
)

// MIMEError describes an error encountered while parsing.
type MIMEError struct {
	Name   string // The name or type of error encountered
	Detail string // Additional detail about the cause of the error, if available
	Severe bool   // Indicates that a portion of the message was lost during parsing
}

func (e *MIMEError) String() string {
	sev := "W"
	if e.Severe {
		sev = "E"
	}
	return fmt.Sprintf("[%s] %s: %s", sev, e.Name, e.Detail)
}
