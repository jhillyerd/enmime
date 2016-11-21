package main

import (
	"fmt"
	"io"
	"net/mail"
	"os"
	"path"
	"strings"

	"github.com/jhillyerd/enmime"
)

// AddressHeaders enumerates SMTP headers that contain email addresses
var addressHeaders = []string{"From", "To", "Delivered-To", "Cc", "Bcc", "Reply-To"}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Missing filename argument")
		os.Exit(1)
	}

	reader, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open file:", err)
		os.Exit(1)
	}

	basename := path.Base(os.Args[1])
	if err = dump(reader, basename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func dump(reader io.Reader, name string) error {
	// Parse message body with enmime
	e, err := enmime.ReadEnvelope(reader)
	if err != nil {
		return fmt.Errorf("During enmime.ReadEnvelope: %v", err)
	}

	h1(name)

	h2("Header")
	for k := range e.Root.Header {
		switch strings.ToLower(k) {
		case "from", "to", "cc", "bcc", "reply-to", "subject":
			continue
		}
		fmt.Printf("    %v: %v\n", k, e.GetHeader(k))
	}
	br()

	h2("Envelope")
	for _, hkey := range addressHeaders {
		addrlist, err := e.AddressList(hkey)
		if err != nil {
			if err == mail.ErrHeaderNotPresent {
				continue
			}
			panic(err)
		}
		fmt.Println("### " + hkey)
		for _, addr := range addrlist {
			fmt.Printf("- %v `<%v>`\n", addr.Name, addr.Address)
		}
		br()
	}
	fmt.Printf("### Subject\n%v\n", e.GetHeader("Subject"))
	br()

	h2("Body Text")
	fmt.Println(e.Text)
	fmt.Println()

	h2("Body HTML")
	fmt.Println(e.HTML)
	fmt.Println()

	h2("Attachment List")
	for _, a := range e.Attachments {
		fmt.Printf("- %v (%v)\n", a.FileName(), a.ContentType())
	}
	fmt.Println()

	h2("MIME Part Tree")
	if e.Root == nil {
		fmt.Println("Message was not MIME encoded")
	} else {
		printPart(e.Root, "    ")
	}

	if len(e.Errors) > 0 {
		br()
		h2("Errors")
		for _, perr := range e.Errors {
			fmt.Println("-", perr)
		}
	}

	return nil
}

func h1(content string) {
	bar := strings.Repeat("-", len(content))
	fmt.Printf("%v\n%v\n\n", content, bar)
}

func h2(content string) {
	fmt.Printf("## %v\n", content)
}

func br() {
	fmt.Println("")
}

// printPart pretty prints the Part tree
func printPart(p *enmime.Part, indent string) {
	sibling := p.NextSibling()
	child := p.FirstChild()

	// Compute indent strings
	myindent := indent + "`-- "
	childindent := indent + "    "
	if sibling != nil {
		myindent = indent + "|-- "
		childindent = indent + "|   "
	}
	if p.Parent() == nil {
		// Root shouldn't be decorated, has no siblings
		myindent = indent
		childindent = indent
	}

	// Format and print this node
	ctype := "MISSING TYPE"
	if p.ContentType() != "" {
		ctype = p.ContentType()
	}
	disposition := ""
	if p.Disposition() != "" {
		disposition = fmt.Sprintf(", disposition: %s", p.Disposition())
	}
	filename := ""
	if p.FileName() != "" {
		filename = fmt.Sprintf(", filename: %q", p.FileName())
	}
	errors := ""
	if len(p.Errors()) > 0 {
		errors = fmt.Sprintf(" (errors: %v)", len(p.Errors()))
	}
	fmt.Printf("%s%s%s%s%s\n", myindent, ctype, disposition, filename, errors)

	// Recurse
	if child != nil {
		printPart(child, childindent)
	}
	if sibling != nil {
		printPart(sibling, indent)
	}
}
