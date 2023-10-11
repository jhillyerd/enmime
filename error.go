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
	// ErrorMalformedChildPart name.
	ErrorMalformedChildPart = "Malformed child part"
)

// MaxPartErrors limits number of part parsing errors, errors after the limit are ignored.
// 0 means unlimited.
//
// Deprecated: This limit may be set via the `MaxStoredPartErrors` Parser option.
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
	maxErrors := MaxPartErrors
	if p.parser != nil && p.parser.maxStoredPartErrors != nil {
		// Override global var.
		maxErrors = *p.parser.maxStoredPartErrors
	}

	if (maxErrors == 0) || (len(p.Errors) < maxErrors) {
		p.Errors = append(p.Errors, err)
	}
}

// ErrorCollector is an interface for collecting errors and warnings during
// parsing.
type ErrorCollector interface {
	AddError(name string, detailFmt string, args ...any)
	AddWarning(name string, detailFmt string, args ...any)
}

type partErrorCollector struct {
	part *Part
}

func (p *partErrorCollector) AddError(name string, detailFmt string, args ...any) {
	p.part.addError(name, detailFmt, args...)
}

func (p *partErrorCollector) AddWarning(name string, detailFmt string, args ...any) {
	p.part.addWarning(name, detailFmt, args...)
}
