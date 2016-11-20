package enmime

import (
	"fmt"
)

type errorName string

const (
	errorBoundaryMissing    errorName = "Boundary Missing"
	errorContentTypeMissing errorName = "Content-Type Missing"
)

// Error describes an error encountered while parsing.
type Error struct {
	Name   string // The name or type of error encountered
	Detail string // Additional detail about the cause of the error, if available
	Severe bool   // Indicates that a portion of the message was lost during parsing
}

// Create a new Error with Severe=false
func newWarning(name errorName, detailFmt string, args ...interface{}) Error {
	return Error{
		string(name),
		fmt.Sprintf(detailFmt, args...),
		false,
	}
}

// Create a new Error with Severe=true
func newError(name errorName, detailFmt string, args ...interface{}) Error {
	return Error{
		string(name),
		fmt.Sprintf(detailFmt, args...),
		true,
	}
}

func (e *Error) String() string {
	sev := "W"
	if e.Severe {
		sev = "E"
	}
	return fmt.Sprintf("[%s] %s: %s", sev, e.Name, e.Detail)
}
