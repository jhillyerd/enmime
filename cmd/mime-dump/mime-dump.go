// Package main outputs a markdown formatted document describing the provided email
package main

import (
	"fmt"
	"os"
	"path"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/cmd"
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

	// basename is used as the markdown title
	basename := path.Base(os.Args[1])
	e, err := enmime.ReadEnvelope(reader)
	if err != nil {
		fmt.Fprintln(os.Stderr, "During enmime.ReadEnvelope:", err)
		os.Exit(1)
	}

	if err = cmd.EnvelopeToMarkdown(os.Stdout, e, basename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
