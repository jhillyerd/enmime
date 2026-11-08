package enmime

import (
	"bytes"
	"net/mail"
	"strings"

	inttp "github.com/jhillyerd/enmime/v2/internal/textproto"
	"github.com/pkg/errors"
)

// sendAddressHeaders are the headers whose values provide envelope recipients for
// Envelope.Send, in delivery order.
var sendAddressHeaders = []string{"To", "Cc", "Bcc"}

// Send encodes the envelope and transmits it via sender, using the address in the "From" header as
// reverse-path, and the addresses in the "To", "Cc" and "Bcc" headers as recipients.
//
// All values of each recipient header are parsed, so a message that folds its recipients across
// several "To:" lines delivers all of them.  Addresses are delivered once each, even when repeated.
//
// "Bcc" addresses are delivered as envelope recipients but the transmitted bytes contain no "Bcc"
// header, per RFC 5322 section 3.6.3 and consistent with MailBuilder.Build().  The header is
// removed from the part tree for the duration of encoding and restored afterwards, so the envelope
// still exposes it to the caller.
//
// Caveats: the message is re-encoded by Part.Encode, which deletes and recomputes each part's
// Content-Transfer-Encoding, and may rewrite charset parameters, so the bytes on the wire can
// differ from the original message.  Callers needing a specific encoding must set
// Part.ContentTransferEncoding (or use WithContentTransferEncoding) before calling Send.
// Part.Epilogue is not written by Encode, so epilogue bytes are dropped.  Sending mutates the part
// tree while encoding, so it is not safe to send the same envelope concurrently.
func (e *Envelope) Send(sender Sender) error {
	return e.SendWithReversePath(sender, "")
}

// SendWithReversePath encodes the envelope and transmits it via sender, using reversePath as the
// envelope reverse-path (bounce address); when reversePath is empty the first address parsed from
// the "From" header is used instead.  Recipients, Bcc handling and the re-encoding caveats are as
// described on Send.
func (e *Envelope) SendWithReversePath(sender Sender, reversePath string) error {
	if e == nil || e.Root == nil {
		return errors.New("cannot send nil Envelope")
	}

	if reversePath == "" {
		var err error
		reversePath, err = e.reversePath()
		if err != nil {
			return err
		}
	}

	recipients, err := e.recipients()
	if err != nil {
		return err
	}

	// Deliver Bcc addresses, but never put them on the wire; see RFC 5322 section 3.6.3.
	// Envelope.header and Root.Header are the same map (see EnvelopeFromPart), so the value is
	// restored on the part's header.
	hadBcc := false
	var savedBcc []string
	bccKey := inttp.CanonicalEmailMIMEHeaderKey("Bcc")
	if bcc, ok := e.Root.Header[bccKey]; ok {
		hadBcc = true
		savedBcc = bcc
		delete(e.Root.Header, bccKey)
	}

	buf := &bytes.Buffer{}
	err = e.Root.Encode(buf)
	if hadBcc {
		e.Root.Header[bccKey] = savedBcc
	}
	if err != nil {
		return err
	}

	return sender.Send(reversePath, recipients, buf.Bytes())
}

// reversePath returns the address portion of the first address in the "From" header.
func (e *Envelope) reversePath() (string, error) {
	addrs, err := e.addressList("From")
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse From header for reverse-path")
	}
	for _, addr := range addrs {
		if addr.Address != "" {
			return addr.Address, nil
		}
	}
	return "", errors.New("no From address to use as reverse-path")
}

// recipients returns the addresses found in the To, Cc and Bcc headers, in that order, de-duplicated
// by address.  Every value of each header is parsed, not just the first.
func (e *Envelope) recipients() ([]string, error) {
	var recips []string
	seen := make(map[string]bool)
	for _, header := range sendAddressHeaders {
		for _, value := range e.headerValues(header) {
			if strings.TrimSpace(value) == "" {
				continue
			}
			addrs, err := ParseAddressList(value)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse %s header for recipients", header)
			}
			for _, addr := range addrs {
				if addr.Address == "" || seen[addr.Address] {
					continue
				}
				seen[addr.Address] = true
				recips = append(recips, addr.Address)
			}
		}
	}
	if len(recips) == 0 {
		return nil, errors.New(ErrorMissingRecipient)
	}
	return recips, nil
}

// addressList parses every value of the named address header, returning the addresses in order.
func (e *Envelope) addressList(header string) ([]*mail.Address, error) {
	var addrs []*mail.Address
	for _, value := range e.headerValues(header) {
		if strings.TrimSpace(value) == "" {
			continue
		}
		parsed, err := ParseAddressList(value)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, parsed...)
	}
	return addrs, nil
}

// headerValues returns the raw values stored for the named header, folding in the canonical key the
// way GetHeaderValues does.
func (e *Envelope) headerValues(header string) []string {
	if e.header == nil {
		return nil
	}
	return (*e.header)[inttp.CanonicalEmailMIMEHeaderKey(header)]
}
