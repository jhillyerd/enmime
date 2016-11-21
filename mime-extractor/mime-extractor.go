package main

import (
	"flag"
	"fmt"
	"github.com/jhillyerd/enmime"
	"io"
	"os"
	"path"
	"strings"
)

var (
	mimefile = flag.String("f", "", "mime(eml) file")
	outdir   = flag.String("o", "", "output dir")
)

func main() {
	flag.Parse()
	if *mimefile == "" {
		fmt.Fprintln(os.Stderr, "Missing filename argument")
		flag.Usage()
		os.Exit(1)
	}
	cwd, _ := os.Getwd()
	if *outdir == "" {
		outdir = &cwd
	}
	fmt.Fprintf(os.Stdout, "Extract attachments in %s\n", *outdir)

	if err := os.MkdirAll(*outdir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Mkdir %s failed.", *outdir)
		os.Exit(2)
	}

	reader, err := os.Open(*mimefile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open file:", err)
		os.Exit(1)
	}

	basename := path.Base(*mimefile)
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

	h2("Envelope")
	fmt.Printf("From: %v  \n", e.GetHeader("From"))
	fmt.Printf("To: %v  \n", e.GetHeader("To"))
	fmt.Printf("Subject: %v  \n", e.GetHeader("Subject"))
	fmt.Println()

	h2("Body Text")
	fmt.Println(e.Text)
	fmt.Println()

	h2("Body HTML")
	fmt.Println(e.HTML)
	fmt.Println()

	h2("Attachment List")
	for _, a := range e.Attachments {
		newFileName := path.Join(*outdir, a.FileName())
		f, err := os.Create(newFileName)
		if err != nil {
			fmt.Printf("Error creating file %q: %v\n", newFileName, err)
		}
		_, err = io.Copy(f, a)
		if err != nil {
			fmt.Printf("Error writing file %q: %v\n", newFileName, err)
		}
		err = f.Close()
		if err != nil {
			fmt.Printf("Error closing file %q: %v\n", newFileName, err)
		}
		fmt.Printf("- %v (%v)\n", a.FileName(), a.ContentType())
	}
	fmt.Println()

	h2("MIME Part Tree")
	if e.Root == nil {
		fmt.Println("Message was not MIME encoded")
	} else {
		printPart(e.Root, "    ")
	}

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
