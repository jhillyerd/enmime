package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/mail"
	"sort"
	"strings"

	"github.com/jhillyerd/enmime"
)

// AddressHeaders enumerates SMTP headers that contain email addresses
var addressHeaders = []string{"From", "To", "Delivered-To", "Cc", "Bcc", "Reply-To"}

// markdown adds some simple HTML tag like methods to a bufio.Writer
type markdown struct {
	*bufio.Writer
}

func (md *markdown) H1(content string) {
	bar := strings.Repeat("=", len(content))
	fmt.Fprintf(md, "%v\n%v\n\n", content, bar)
}

func (md *markdown) H2(content string) {
	fmt.Fprintf(md, "## %v\n", content)
}

func (md *markdown) H3(content string) {
	fmt.Fprintf(md, "### %v\n", content)
}

// Printf implements fmt.Printf for markdown
func (md *markdown) Printf(format string, args ...interface{}) {
	fmt.Fprintf(md, format, args...)
}

// Println implements fmt.Println for markdown
func (md *markdown) Println(args ...interface{}) {
	fmt.Fprintln(md, args...)
}

// EnvelopeToMarkdown renders the contents of an enmime.Envelope in Markdown format. Used by
// mime-dump and mime-extractor commands.
func EnvelopeToMarkdown(w io.Writer, e *enmime.Envelope, name string) error {
	md := &markdown{bufio.NewWriter(w)}

	md.H1(name)

	// Output a sorted list of headers, minus the ones displayed later
	md.H2("Header")
	if e.Root != nil && e.Root.Header != nil {
		keys := make([]string, 0, len(e.Root.Header))
		for k := range e.Root.Header {
			switch strings.ToLower(k) {
			case "from", "to", "cc", "bcc", "reply-to", "subject":
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			md.Printf("    %v: %v\n", k, e.GetHeader(k))
		}
	}
	md.Println()

	md.H2("Envelope")
	for _, hkey := range addressHeaders {
		addrlist, err := e.AddressList(hkey)
		if err != nil {
			if err == mail.ErrHeaderNotPresent {
				continue
			}
			return err
		}
		md.H3(hkey)
		for _, addr := range addrlist {
			md.Printf("- %v `<%v>`\n", addr.Name, addr.Address)
		}
		md.Println()
	}
	md.H3("Subject")
	md.Println(e.GetHeader("Subject"))
	md.Println()

	md.H2("Body Text")
	md.Println(e.Text)
	md.Println()

	md.H2("Body HTML")
	md.Println(e.HTML)
	md.Println()

	md.H2("Attachment List")
	for _, a := range e.Attachments {
		md.Printf("- %v (%v)\n", a.FileName, a.ContentType)
	}
	md.Println()

	md.H2("MIME Part Tree")
	if e.Root == nil {
		md.Println("Message was not MIME encoded")
	} else {
		FormatPart(md, e.Root, "    ")
	}

	if len(e.Errors) > 0 {
		md.Println()
		md.H2("Errors")
		for _, perr := range e.Errors {
			md.Println("-", perr)
		}
	}

	return md.Flush()
}

// FormatPart pretty prints the Part tree
func FormatPart(w io.Writer, p *enmime.Part, indent string) {
	if p == nil {
		return
	}

	sibling := p.NextSibling
	child := p.FirstChild

	// Compute indent strings
	myindent := indent + "`-- "
	childindent := indent + "    "
	if sibling != nil {
		myindent = indent + "|-- "
		childindent = indent + "|   "
	}
	if p.Parent == nil {
		// Root shouldn't be decorated, has no siblings
		myindent = indent
		childindent = indent
	}

	// Format and print this node
	ctype := "MISSING TYPE"
	if p.ContentType != "" {
		ctype = p.ContentType
	}
	disposition := ""
	if p.Disposition != "" {
		disposition = fmt.Sprintf(", disposition: %s", p.Disposition)
	}
	filename := ""
	if p.FileName != "" {
		filename = fmt.Sprintf(", filename: %q", p.FileName)
	}
	errors := ""
	if len(p.Errors) > 0 {
		errors = fmt.Sprintf(" (errors: %v)", len(p.Errors))
	}
	fmt.Fprintf(w, "%s%s%s%s%s\n", myindent, ctype, disposition, filename, errors)

	// Recurse
	FormatPart(w, child, childindent)
	FormatPart(w, sibling, indent)
}
