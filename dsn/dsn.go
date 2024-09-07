// Package dsn meant to work with Delivery Status Notification (DSN) per rfc3464:
// https://datatracker.ietf.org/doc/html/rfc3464
package dsn

import (
	"strings"

	"github.com/jhillyerd/enmime/v2/internal/textproto"
)

// Report represents delivery status report as per https://datatracker.ietf.org/doc/html/rfc6522.
// It is container of the MIME type multipart/report and contains 3 parts.
type Report struct {
	// Explanation contains a human-readable description of the condition(s) that caused the report to be generated.
	Explanation Explanation
	// DeliveryStatus provides a machine-readable description of the condition(s) that caused the report to be generated.
	DeliveryStatus DeliveryStatus
	// OriginalMessage is optional original message or its portion.
	OriginalMessage []byte
}

// Explanation contains a human-readable description of the condition(s) that caused the report to be generated.
// Where a description of the error is desired in several languages or several media, a multipart/alternative construct MAY be used.
type Explanation struct {
	// Text is a text/plain part of the explanation.
	Text string
	// HTML is a text/html part of the explanation.
	HTML string
}

// DeliveryStatus provides a machine-readable description of the condition(s) that caused the report to be generated.
type DeliveryStatus struct {
	// MessageDSN is Delivery Status Notification per message.
	MessageDSNs []textproto.MIMEHeader
	// RecipientDSN is Delivery Status Notification per recipient.
	RecipientDSNs []textproto.MIMEHeader
}

// IsFailed returns true if Action field is "failed", meaning message could not be delivered to the recipient.
func IsFailed(n textproto.MIMEHeader) bool {
	return strings.EqualFold(n.Get("Action"), "failed")
}
