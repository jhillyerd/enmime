package cmd_test

import (
	"log"
	"os"
	"strings"

	"github.com/jhillyerd/enmime/v2"
	"github.com/jhillyerd/enmime/v2/cmd"
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
--Enmime-Test-100
Content-Transfer-Encoding: base64
Content-Disposition: inline;
	filename=favicon.png
Content-Type: image/png;
	x-unix-mode=0644;
	name="favicon.png"
Content-Id: <8B8481A2-25CA-4886-9B5A-8EB9115DD064@skynet>

iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJ
bWFnZVJlYWR5ccllPAAAAlFJREFUeNqUU8tOFEEUPVVdNV3dPe8xYRBnjGhmBgKjKzCIiQvBoIaN
bly5Z+PSv3Aj7DSiP2B0rwkLGVdGgxITSCRIJGSMEQWZR3eVt5sEFBgTb/dN1yvnnHtPNTPG4Pqd
HgCMXnPRSZrpSuH8vUJu4DE4rYHDGAZDX62BZttHqTiIayM3gGiXQsgYLEvATaqxU+dy1U13YXap
XptpNHY8iwn8KyIAzm1KBdtRZWErpI5lEWTXp5Z/vHpZ3/wyKKwYGGOdAYwR0EZwoezTYApBEIOb
yELl/aE1/83cp40Pt5mxqCKrE4Ck+mVWKKcI5tA8BLEhRBKJLjez6a7MLq7XZtp+yyOawwCBtkiB
VZDKzRk4NN7NQBMYPHiZDFhXY+p9ff7F961vVcnl4R5I2ykJ5XFN7Ab7Gc61VoipNBKF+PDyztu5
lfrSLT/wIwCxq0CAGtXHZTzqR2jtwQiXONma6hHpj9sLT7YaPxfTXuZdBGA02Wi7FS48YiTfj+i2
NhqtdhP5RC8mh2/Op7y0v6eAcWVLFT8D7kWX5S9mepp+C450MV6aWL1cGnvkxbwHtLW2B9AOkLeU
d9KEDuh9fl/7CEj7YH5g+3r/lWfF9In7tPz6T4IIwBJOr1SJyIGQMZQbsh5P9uBq5VJtqHh2mo49
pdw5WFoEwKWqWHacaWOjQXWGcifKo6vj5RGS6zykI587XeUIQDqJSmAp+lE4qt19W5P9o8+Lma5D
cjsC8JiT607lMVkdqQ0Vyh3lHhmh52tfNy78ajXv0rgYzv8nfwswANuk+7sD/Q0aAAAAAElFTkSu
QmCC
--Enmime-Test-100
Content-Transfer-Encoding: base64
Content-Type: text/html; name="test.html"
Content-Disposition: attachment; filename=test.html

PGh0bWw+Cg==
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
	// - test.html (text/html)
	//
	// ## Inline List
	// - favicon.png (image/png)
	//   Content-ID: 8B8481A2-25CA-4886-9B5A-8EB9115DD064@skynet
	//
	// ## Other Part List
	//
	// ## MIME Part Tree
	//     multipart/mixed
	//     |-- text/plain
	//     |-- text/html
	//     |-- image/png, disposition: inline, filename: "favicon.png"
	//     `-- text/html, disposition: attachment, filename: "test.html"
	//
}
