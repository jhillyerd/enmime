package enmime

import "net/smtp"

// Sender provides a method for enmime to send an email.
type Sender interface {
	Send(from string, to []string, msg []byte) error
}

// SMTPSender is a Sender backed by Go's built-in net/smtp.SendMail function.
type SMTPSender struct {
	addr string
	auth smtp.Auth
}

var _ Sender = &SMTPSender{}

// NewSMTP creates a new SMTPSender.  If no authentication is required, `auth` may be nil.
func NewSMTP(addr string, auth smtp.Auth) *SMTPSender {
	return &SMTPSender{addr, auth}
}

// Send a message using net/smtp.SendMail.
func (s *SMTPSender) Send(from string, to []string, msg []byte) error {
	return smtp.SendMail(s.addr, s.auth, from, to, msg)
}
