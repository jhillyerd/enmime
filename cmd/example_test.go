package cmd_test

import (
	"log"
	"os"
	"strings"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/cmd"
)

func Example() {
	mail := `From: James Hillyerd <james@inbucket.org>
To: Greg Reader <greg@inbucket.org>, Root Node <root@inbucket.org>
Date: Sat, 04 Dec 2016 18:38:25 -0800
Subject: Example Message
Content-Type: multipart/mixed; boundary="Enmime-Test-100"

--Enmime-Test-100
Content-Type: text/plain

Text section.
--Enmime-Test-100
Content-Type: text/html

<em>HTML</em> section.
--Enmime-Test-100--
`
	// Convert MIME text to Envelope
	r := strings.NewReader(mail)
	env, err := enmime.ReadEnvelope(r)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.EnvelopeToMarkdown(os.Stdout, env, "Example Message Output")
	if err != nil {
		log.Fatal(err)
		return
	}

	// Output:
	// Example Message Output
	// ======================
	//
	// ## Header
	//     Content-Type: multipart/mixed; boundary="Enmime-Test-100"
	//     Date: Sat, 04 Dec 2016 18:38:25 -0800
	//
	// ## Envelope
	// ### From
	// - James Hillyerd `<james@inbucket.org>`
	//
	// ### To
	// - Greg Reader `<greg@inbucket.org>`
	// - Root Node `<root@inbucket.org>`
	//
	// ### Subject
	// Example Message
	//
	// ## Body Text
	// Text section.
	//
	// ## Body HTML
	// <em>HTML</em> section.
	//
	// ## Attachment List
	//
	// ## MIME Part Tree
	//     multipart/mixed
	//     |-- text/plain
	//     `-- text/html
	//
}
