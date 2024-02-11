package enmime_test

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jhillyerd/enmime"
)

// ExampleBuilder illustrates how to build and send a MIME encoded message.
func ExampleBuilder() {
	// Create an `SMTPSender` which relies on Go's built-in net/smtp package. Advanced users may
	// provide their own implementation of `Sender`, or mock it in unit tests.
	// For example:
	//
	// smtpHost := "smtp.relay.host:25"
	// smtpAuth := smtp.PlainAuth("", "user", "pw", "host")
	// sender := enmime.NewSMTP(smtpHost, smtpAuth)

	// Instead, we use a fake sender which prints to stdout:
	sender := &stdoutSender{}

	// MailBuilder is (mostly) immutable, each method below returns a new MailBuilder without
	// modifying the original.
	master := enmime.Builder().
		From("Do Not Reply", "noreply@inbucket.org").
		Subject("Inbucket Newsletter").
		Text([]byte("Text body")).
		HTML([]byte("<p>HTML body</p>"))

	// Force stable output for testing; not needed in production.
	master = master.RandSeed(1).Date(time.Date(2024, 1, 1, 13, 14, 15, 16, time.UTC))

	msg := master.To("Esteemed Customer", "user1@inbucket.org")
	err := msg.Send(sender)
	if err != nil {
		panic(err)
	}

	msg = master.To("Another Customer", "user2@inbucket.org")
	err = msg.Send(sender)
	if err != nil {
		panic(err)
	}

	// Output:
	// MAIL FROM:<noreply@inbucket.org>
	// RCPT TO:<user1@inbucket.org>
	// DATA
	// Content-Type: multipart/alternative;
	//  boundary=enmime-52fdfc07-2182-454f-963f-5f0f9a621d72
	// Date: Mon, 01 Jan 2024 13:14:15 +0000
	// From: "Do Not Reply" <noreply@inbucket.org>
	// Mime-Version: 1.0
	// Subject: Inbucket Newsletter
	// To: "Esteemed Customer" <user1@inbucket.org>
	//
	// --enmime-52fdfc07-2182-454f-963f-5f0f9a621d72
	// Content-Type: text/plain; charset=utf-8
	//
	// Text body
	// --enmime-52fdfc07-2182-454f-963f-5f0f9a621d72
	// Content-Type: text/html; charset=utf-8
	//
	// <p>HTML body</p>
	// --enmime-52fdfc07-2182-454f-963f-5f0f9a621d72--
	//
	// MAIL FROM:<noreply@inbucket.org>
	// RCPT TO:<user2@inbucket.org>
	// DATA
	// Content-Type: multipart/alternative;
	//  boundary=enmime-037c4d7b-bb04-47d1-a2c6-4981855ad868
	// Date: Mon, 01 Jan 2024 13:14:15 +0000
	// From: "Do Not Reply" <noreply@inbucket.org>
	// Mime-Version: 1.0
	// Subject: Inbucket Newsletter
	// To: "Another Customer" <user2@inbucket.org>
	//
	// --enmime-037c4d7b-bb04-47d1-a2c6-4981855ad868
	// Content-Type: text/plain; charset=utf-8
	//
	// Text body
	// --enmime-037c4d7b-bb04-47d1-a2c6-4981855ad868
	// Content-Type: text/html; charset=utf-8
	//
	// <p>HTML body</p>
	// --enmime-037c4d7b-bb04-47d1-a2c6-4981855ad868--
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
	fmt.Println(p3error.Error())
	fmt.Println()

	// All Part errors are collected and placed into Envelope.Errors
	fmt.Println("Envelope errors:")
	for _, e := range env.Errors {
		fmt.Println(e.Error())
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

func ExampleEnvelope_GetHeaderKeys() {
	// Open a sample message file.
	r, err := os.Open("testdata/mail/qp-utf8-header.raw")
	if err != nil {
		fmt.Print(err)
		return
	}

	// Parse message with enmime.
	env, err := enmime.ReadEnvelope(r)
	if err != nil {
		fmt.Print(err)
		return
	}

	// A list of headers is retrieved via Envelope.GetHeaderKeys().
	headers := env.GetHeaderKeys()
	sort.Strings(headers)

	// Print each header, key and value.
	for _, header := range headers {
		fmt.Printf("%s: %v\n", header, env.GetHeader(header))
	}

	// Output:
	// Content-Type: multipart/alternative; boundary="------------020203040006070307010003"
	// Date: Fri, 19 Oct 2012 12:22:49 -0700
	// From: James Hillyerd <jamehi03@jamehi03lx.noa.com>, André Pirard <PIRARD@vm1.ulg.ac.be>
	// Message-Id: <5081A889.3020108@jamehi03lx.noa.com>
	// Mime-Version: 1.0
	// Sender: André Pirard <PIRARD@vm1.ulg.ac.be>
	// Subject: MIME UTF8 Test ¢ More Text
	// To: "Mirosław Marczak" <marczak@inbucket.com>
	// User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64; rv:16.0) Gecko/20121010 Thunderbird/16.0.1
}

type stdoutSender struct{}

func (s *stdoutSender) Send(from string, tos []string, msg []byte) error {
	fmt.Printf("MAIL FROM:<%v>\n", from)
	for _, to := range tos {
		fmt.Printf("RCPT TO:<%v>\n", to)
	}

	fmt.Println("DATA")
	lines := bytes.Split(msg, []byte{'\r'})
	for _, line := range lines {
		line = bytes.Trim(line, "\r\n")
		fmt.Println(string(line))
	}

	return nil
}
