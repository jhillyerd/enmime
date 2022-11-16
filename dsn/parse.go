package dsn

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/textproto"

	"github.com/jhillyerd/enmime"
)

// ParseReport parses p as a "container" for delivery status report (per rfc6522) if p is "multipart/report".
// Otherwise returns nil.
func ParseReport(p *enmime.Part) (*Report, error) {
	if !isMultipartReport(p.ContentType) {
		return nil, nil
	}

	var report Report
	// first part is explanation
	explanation := p.FirstChild
	if explanation == nil || !setExplanation(&report.Explanation, explanation) {
		return &report, nil
	}

	// second part is delivery-status
	deliveryStatus := explanation.NextSibling
	if deliveryStatus == nil || !isDeliveryStatus(deliveryStatus.ContentType) {
		return &report, nil
	}

	ds, err := parseDeliveryStatus(deliveryStatus.Content)
	if err != nil {
		return nil, err
	}
	report.DeliveryStatus = ds

	// third part is original email
	originalEmail := deliveryStatus.NextSibling
	if originalEmail == nil || !isEmail(originalEmail.ContentType) {
		return &report, nil
	}

	report.OriginalMessage = originalEmail.Content

	return &report, nil
}

func isMultipartReport(ct string) bool {
	return ct == "multipart/report"
}

func isDeliveryStatus(ct string) bool {
	return ct == "message/delivery-status" || ct == "message/global-delivery-status"
}

func isEmail(ct string) bool {
	return ct == "message/rfc822"
}

func setExplanation(e *Explanation, p *enmime.Part) bool {
	if p == nil {
		return false
	}

	switch p.ContentType {
	case "text/plain", "": // treat no content-type as text
		e.Text = string(p.Content)
		return true
	case "text/html":
		e.HTML = string(p.Content)
		return true
	case "multipart/alternative":
		// the structure is next:
		// 	multipart/alternative
		// 	- text/plain (FirstChild)
		// 	- text/html (FirstChild.NextSibling)
		if setExplanation(e, p.FirstChild) {
			return setExplanation(e, p.FirstChild.NextSibling)
		}
	default:
		return false
	}

	return false
}

func parseDeliveryStatus(date []byte) (DeliveryStatus, error) {
	fields, err := parseDeliveryStatusMessage(date)
	if err != nil {
		return DeliveryStatus{}, fmt.Errorf("parse delivey status: %w", err)
	}

	perMessage, perRecipient := splitDSNFields(fields)

	return DeliveryStatus{
		MessageDSNs:   perMessage,
		RecipientDSNs: perRecipient,
	}, nil
}

// parseDeliveryStatusMessage parses delivery-status message per https://www.rfc-editor.org/rfc/rfc3464#section-2.1
// The body of a message/delivery-status consists of one or more "fields" formatted according to the ABNF of
// RFC 822 header "fields". In other words, body of delivery status is multiple headers separated by blank line.
// First part is per-message fields, following by per-recipient fields.
func parseDeliveryStatusMessage(data []byte) ([]textproto.MIMEHeader, error) {
	if len(data) > 0 && data[len(data)-1] != '\n' { // additional new line if missing
		data = append(data, byte('\n'))
	}
	data = append(data, byte('\n')) // fix for https://github.com/golang/go/issues/47897 - can't read messages with headers only

	r := textproto.NewReader(bufio.NewReader(bytes.NewReader(data)))
	var (
		fields []textproto.MIMEHeader
		err    error
	)
	// parse body as multiple header fields separated by blank lines
	for err == nil {
		h, err := r.ReadMIMEHeader()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("read MIME header fields: %w", err)
		}

		fields = append(fields, h)
	}

	return fields, nil
}

// splitDSNFields splits into per-message and per-recipient fields.
// First part is per-message fields, following by per-recipient fields.
func splitDSNFields(fields []textproto.MIMEHeader) (perMessage, perRecipient []textproto.MIMEHeader) {
	for i, f := range fields {
		if isPerMessageDSN(f) {
			perMessage = append(perMessage, f)
			continue
		}

		perRecipient = fields[i:]
		break
	}

	return
}

// isPerMessageDSN returns true if field is per-message DSN field.
// According to https://datatracker.ietf.org/doc/html/rfc3464#section-3, minimal per-message DSN must have Reporting-MTA field.
func isPerMessageDSN(header textproto.MIMEHeader) bool {
	return header.Get("Reporting-MTA") != ""
}
