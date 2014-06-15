package main

import (
	"fmt"
	"io"
	"net/mail"
	"os"
	"path"
	"strings"

	"github.com/jhillyerd/go.enmime"
)

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
	// Read email using Go's net/mail
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		return fmt.Errorf("During mail.ReadMessage: %v", err)
	}

	// Parse message body with enmime
	mime, err := enmime.ParseMIMEBody(msg)
	if err != nil {
		return fmt.Errorf("During enmime.ParseMIMEBody: %v", err)
	}

	h1(name)

	h2("Envelope")
	fmt.Printf("From: %v  \n", mime.GetHeader("From"))
	fmt.Printf("To: %v  \n", mime.GetHeader("To"))
	fmt.Printf("Subject: %v  \n", mime.GetHeader("Subject"))
	fmt.Println()

	h2("Body Text")
	fmt.Println(mime.Text)
	fmt.Println()

	h2("Body HTML")
	fmt.Println(mime.Html)
	fmt.Println()

	h2("Attachment List")
	for _, a := range mime.Attachments {
		fmt.Printf("- %v (%v)\n", a.FileName(), a.ContentType())
	}
	fmt.Println()

	h2("MIME Part Tree")
	printPart(mime.Root, "    ")

	return nil
}

func h1(content string) {
	bar := strings.Repeat("=", len(content))
	fmt.Printf("%v\n%v\n\n", content, bar)
}

func h2(content string) {
	bar := strings.Repeat("-", len(content))
	fmt.Printf("%v\n%v\n", content, bar)
}

// printPart pretty prints the MIMEPart tree
func printPart(p enmime.MIMEPart, indent string) {
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
		disposition = ", disposition: " + p.Disposition()
	}
	filename := ""
	if p.FileName() != "" {
		filename = ", filename: \"" + p.FileName() + "\""
	}
	fmt.Printf("%s%s%s%s\n", myindent, ctype, disposition, filename)

	// Recurse
	if child != nil {
		printPart(child, childindent)
	}
	if sibling != nil {
		printPart(sibling, indent)
	}
}
