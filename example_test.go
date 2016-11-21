package enmime_test

import (
	"fmt"
	"os"

	"github.com/jhillyerd/enmime"
)

func Example() {
	r, _ := os.Open("testdata/mail/qp-utf8-header.raw")
	env, _ := enmime.ReadEnvelope(r) // Parse message body with enmime

	// Headers can be retrieved via Envelope.GetHeader(name)
	fmt.Printf("From: %v\n", env.GetHeader("From"))

	// Address-type headers can be parsed into a list of decoded mail.Address structs
	alist, _ := env.AddressList("To")
	for _, addr := range alist {
		fmt.Printf("To: %s <%s>\n", addr.Name, addr.Address)
	}

	// enmime can decode quoted-printable headers
	fmt.Printf("Subject: %v\n", env.GetHeader("Subject"))

	// The plain text body is available as mime.Text
	fmt.Printf("Text Body: %v chars\n", len(env.Text))

	// The HTML body is stored in mime.HTML
	fmt.Printf("HTML Body: %v chars\n", len(env.HTML))

	// mime.Inlines is a slice of inlined attacments
	fmt.Printf("Inlines: %v\n", len(env.Inlines))

	// mime.Attachments contains the non-inline attachments
	fmt.Printf("Attachments: %v\n", len(env.Attachments))

	// Output:
	// From: James Hillyerd <jamehi03@jamehi03lx.noa.com>
	// To: Mirosław Marczak <marczak@inbucket.com>
	// Subject: MIME UTF8 Test ¢ More Text
	// Text Body: 1300 chars
	// HTML Body: 1736 chars
	// Inlines: 0
	// Attachments: 0
}
