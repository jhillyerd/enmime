package cmd

import (
	"fmt"
	"io"

	"github.com/jhillyerd/enmime"
)

// printPart pretty prints the Part tree
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
