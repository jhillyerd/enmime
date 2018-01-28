package enmime_test

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/jhillyerd/enmime"
)

// ExampleBuilder illustrates how to build and send a MIME encoded message.
func ExampleBuilder() {
	smtpHost := "smtp.relay.host:25"
	smtpAuth := smtp.PlainAuth("", "user", "pw", "host")

	// MailBuilder is (mostly) immutable, each method below returns a new MailBuilder without
	// modifying the original.
	master := enmime.Builder().
		From("Do Not Reply", "noreply@inbucket.org").
		Subject("Inbucket Newsletter").
		Text([]byte("Text body")).
		HTML([]byte("<p>HTML body</p>"))

	// master is immutable, causing each msg below to have a single recipient.
	msg := master.To("Esteemed Customer", "user1@inbucket.org")
	msg.Send(smtpHost, smtpAuth)

	msg = master.To("Another Customer", "user2@inbucket.org")
	msg.Send(smtpHost, smtpAuth)
}

func ExampleReadEnvelope() {
	// Open a sample message file.
	r, err := os.Open("testdata/mail/qp-utf8-header.raw")
	if err != nil {
		fmt.Print(err)
		return
	}

	// Parse message body with enmime.
	env, err := enmime.ReadEnvelope(r)
	if err != nil {
		fmt.Print(err)
		return
	}

	// Headers can be retrieved via Envelope.GetHeader(name).
	fmt.Printf("From: %v\n", env.GetHeader("From"))

	// Address-type headers can be parsed into a list of decoded mail.Address structs.
	alist, _ := env.AddressList("To")
	for _, addr := range alist {
		fmt.Printf("To: %s <%s>\n", addr.Name, addr.Address)
	}

	// enmime can decode quoted-printable headers.
	fmt.Printf("Subject: %v\n", env.GetHeader("Subject"))

	// The plain text body is available as mime.Text.
	fmt.Printf("Text Body: %v chars\n", len(env.Text))

	// The HTML body is stored in mime.HTML.
	fmt.Printf("HTML Body: %v chars\n", len(env.HTML))

	// mime.Inlines is a slice of inlined attacments.
	fmt.Printf("Inlines: %v\n", len(env.Inlines))

	// mime.Attachments contains the non-inline attachments.
	fmt.Printf("Attachments: %v\n", len(env.Attachments))

	// Output:
	// From: James Hillyerd <jamehi03@jamehi03lx.noa.com>, André Pirard <PIRARD@vm1.ulg.ac.be>
	// To: Mirosław Marczak <marczak@inbucket.com>
	// Subject: MIME UTF8 Test ¢ More Text
	// Text Body: 1300 chars
	// HTML Body: 1736 chars
	// Inlines: 0
	// Attachments: 0
}

// ExampleEnvelope demonstrates the relationship between Envelope and Parts.
func ExampleEnvelope() {
	// Create sample message in memory
	raw := `From: user@inbucket.org
Subject: Example message
Content-Type: multipart/alternative; boundary=Enmime-100

--Enmime-100
Content-Type: text/plain
X-Comment: part1

hello!
--Enmime-100
Content-Type: text/html
X-Comment: part2

<b>hello!</b>
--Enmime-100
Content-Type: text/plain
Content-Disposition: attachment;
filename=hi.txt
X-Comment: part3

hello again!
--Enmime-100--
`

	// Parse message body with enmime.ReadEnvelope
	r := strings.NewReader(raw)
	env, err := enmime.ReadEnvelope(r)
	if err != nil {
		fmt.Print(err)
		return
	}

	// The root Part contains the message header, which is also available via the
	// Envelope.GetHeader() method.
	fmt.Printf("Root Part Subject: %q\n", env.Root.Header.Get("Subject"))
	fmt.Printf("Envelope Subject: %q\n", env.GetHeader("Subject"))
	fmt.Println()

	// The text from part1 is consumed and placed into the Envelope.Text field.
	fmt.Printf("Text Content: %q\n", env.Text)

	// But part1 is also available as a child of the root Part.  Only the headers may be accessed,
	// because the content has been consumed.
	part1 := env.Root.FirstChild
	fmt.Printf("Part 1 X-Comment: %q\n", part1.Header.Get("X-Comment"))
	fmt.Println()

	// The HTML from part2 is consumed and placed into the Envelope.HTML field.
	fmt.Printf("HTML Content: %q\n", env.HTML)

	// And part2 is available as the second child of the root Part. Only the headers may be
	// accessed, because the content has been consumed.
	part2 := env.Root.FirstChild.NextSibling
	fmt.Printf("Part 2 X-Comment: %q\n", part2.Header.Get("X-Comment"))
	fmt.Println()

	// Because part3 has a disposition of attachment, it is added to the Envelope.Attachments
	// slice
	fmt.Printf("Attachment 1 X-Comment: %q\n", env.Attachments[0].Header.Get("X-Comment"))

	// And is still available as the third child of the root Part
	part3 := env.Root.FirstChild.NextSibling.NextSibling
	fmt.Printf("Part 3 X-Comment: %q\n", part3.Header.Get("X-Comment"))

	// The content of Attachments, Inlines and OtherParts are available as a slice of bytes
	fmt.Printf("Part 3 Content: %q\n", part3.Content)

	// part3 contained a malformed header line, enmime has attached an Error to it
	p3error := part3.Errors[0]
	fmt.Println(p3error.String())
	fmt.Println()

	// All Part errors are collected and placed into Envelope.Errors
	fmt.Println("Envelope errors:")
	for _, e := range env.Errors {
		fmt.Println(e.String())
	}

	// Output:
	// Root Part Subject: "Example message"
	// Envelope Subject: "Example message"
	//
	// Text Content: "hello!"
	// Part 1 X-Comment: "part1"
	//
	// HTML Content: "<b>hello!</b>"
	// Part 2 X-Comment: "part2"
	//
	// Attachment 1 X-Comment: "part3"
	// Part 3 X-Comment: "part3"
	// Part 3 Content: "hello again!"
	// [W] Malformed Header: Continued line "filename=hi.txt" was not indented
	//
	// Envelope errors:
	// [W] Malformed Header: Continued line "filename=hi.txt" was not indented
}
