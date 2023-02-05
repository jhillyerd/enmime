package enmime_test

import (
	"net/smtp"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime"
)

func TestSMTPSend(t *testing.T) {
	from := "user@example.com"
	auth := smtp.PlainAuth("", from, "password", "mail.example.com")
	s := enmime.NewSMTP("0.0.0.0", auth)

	// Satisfy requirements of the Send method and use an intentionally malformed From Address to
	// elicit an expected error from smtp.SendMail, which can be type-checked and verified.
	err := s.Send(from+"\rinvalid", []string{"to@example.com"}, []byte("message content"))
	if err == nil {
		t.Fatal("Send() returned nil error, wanted one.")
	}
	if !strings.Contains(err.Error(), "smtp: A line must not contain CR or LF") {
		t.Fatalf("Send() did not return expected error, failed: %s", err.Error())
	}
}
