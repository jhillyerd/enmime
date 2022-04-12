package enmime

import (
	"fmt"
)

const (
	// ErrorMalformedBase64 name.
	ErrorMalformedBase64 = "Malformed Base64"
	// ErrorMalformedHeader name.
	ErrorMalformedHeader = "Malformed Header"
	// ErrorMissingBoundary name.
	ErrorMissingBoundary = "Missing Boundary"
	// ErrorMissingContentType name.
	ErrorMissingContentType = "Missing Content-Type"
	// ErrorCharsetConversion name.
	ErrorCharsetConversion = "Character Set Conversion"
	// ErrorContentEncoding name.
	ErrorContentEncoding = "Content Encoding"
	// ErrorPlainTextFromHTML name.
	ErrorPlainTextFromHTML = "Plain Text from HTML"
	// ErrorCharsetDeclaration name.
	ErrorCharsetDeclaration = "Character Set Declaration Mismatch"
	// ErrorMissingRecipient name.
	ErrorMissingRecipient = "no recipients (to, cc, bcc) set"
)

// MaxPartErrors limits number of part parsing errors, errors after the limit are ignored. 0 means unlimited.
var MaxPartErrors = 0

// Error describes an error encountered while parsing.
type Error struct {
	Name   string // The name or type of error encountered, from Error consts.
	Detail string // Additional detail about the cause of the error, if available.
	Severe bool   // Indicates that a portion of the message was lost during parsing.
}

// Error formats the enmime.Error as a string.
func (e *Error) Error() string {
	sev := "W"
	if e.Severe {
		sev = "E"
	}
	return fmt.Sprintf("[%s] %s: %s", sev, e.Name, e.Detail)
}

// String formats the enmime.Error as a string. DEPRECATED; use Error() instead.
func (e *Error) String() string {
	return e.Error()
}

// addWarning builds a severe Error and appends to the Part error slice.
func (p *Part) addError(name string, detailFmt string, args ...interface{}) {
	p.addProblem(&Error{
		name,
		fmt.Sprintf(detailFmt, args...),
		true,
	})
}

// addWarning builds a non-severe Error and appends to the Part error slice.
func (p *Part) addWarning(name string, detailFmt string, args ...interface{}) {
	p.addProblem(&Error{
		name,
		fmt.Sprintf(detailFmt, args...),
		false,
	})
}

// addProblem adds general *Error to the Part error slice.
func (p *Part) addProblem(err *Error) {
	if (MaxPartErrors == 0) || (len(p.Errors) < MaxPartErrors) {
		p.Errors = append(p.Errors, err)
	}
}
