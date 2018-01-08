package enmime

import (
	"fmt"
)

const (
	// ErrorMalformedBase64 name
	ErrorMalformedBase64 = "Malformed Base64"
	// ErrorMalformedHeader name
	ErrorMalformedHeader = "Malformed Header"
	// ErrorMissingBoundary name
	ErrorMissingBoundary = "Missing Boundary"
	// ErrorMissingContentType name
	ErrorMissingContentType = "Missing Content-Type"
	// ErrorCharsetConversion name
	ErrorCharsetConversion = "Character Set Conversion"
	// ErrorContentEncoding name
	ErrorContentEncoding = "Content Encoding"
	// ErrorPlainTextFromHTML name
	ErrorPlainTextFromHTML = "Plain Text from HTML"
)

// Error describes an error encountered while parsing.
type Error struct {
	Name   string // The name or type of error encountered, from Error consts
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
func (p *Part) addError(name string, detailFmt string, args ...interface{}) {
	p.Errors = append(
		p.Errors,
		Error{
			name,
			fmt.Sprintf(detailFmt, args...),
			true,
		})
}

// addWarning builds a non-severe Error and appends to the Part error slice
func (p *Part) addWarning(name string, detailFmt string, args ...interface{}) {
	p.Errors = append(
		p.Errors,
		Error{
			name,
			fmt.Sprintf(detailFmt, args...),
			false,
		})
}
