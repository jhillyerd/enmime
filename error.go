package enmime

import (
	"fmt"
)

type errorName string

const (
	errorMalformedHeader    errorName = "Malformed Header"
	errorMissingBoundary    errorName = "Missing Boundary"
	errorMissingContentType errorName = "Missing Content-Type"
	errorCharsetConversion  errorName = "Character Set Conversion"
	errorContentEncoding    errorName = "Content Encoding"
	errorPlainTextFromHTML  errorName = "Plain Text from HTML"
)

// Error describes an error encountered while parsing.
type Error struct {
	Name   string // The name or type of error encountered
	Detail string // Additional detail about the cause of the error, if available
	Severe bool   // Indicates that a portion of the message was lost during parsing
}

// String formats the enmime.Error as a string
func (e *Error) String() string {
	sev := "W"
	if e.Severe {
		sev = "E"
	}
	return fmt.Sprintf("[%s] %s: %s", sev, e.Name, e.Detail)
}

// addWarning builds a severe Error and appends to the Part error slice
func (p *Part) addError(name errorName, detailFmt string, args ...interface{}) {
	p.Errors = append(
		p.Errors,
		Error{
			string(name),
			fmt.Sprintf(detailFmt, args...),
			true,
		})
}

// addWarning builds a non-severe Error and appends to the Part error slice
func (p *Part) addWarning(name errorName, detailFmt string, args ...interface{}) {
	p.Errors = append(
		p.Errors,
		Error{
			string(name),
			fmt.Sprintf(detailFmt, args...),
			false,
		})
}
