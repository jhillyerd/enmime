// Package main extracts attachments from the provided email
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/cmd"
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

	if err := os.MkdirAll(*outdir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Mkdir %s failed.", *outdir)
		os.Exit(2)
	}

	reader, err := os.Open(*mimefile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open file:", err)
		os.Exit(1)
	}

	// basename is used as the markdown title
	basename := path.Base(*mimefile)
	e, err := enmime.ReadEnvelope(reader)
	if err != nil {
		fmt.Fprintln(os.Stderr, "During enmime.ReadEnvelope:", err)
		os.Exit(1)
	}

	if err = cmd.EnvelopeToMarkdown(os.Stdout, e, basename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Write out attachments
	fmt.Fprintf(os.Stderr, "\nExtracting attachments into %s...", *outdir)
	for _, a := range e.Attachments {
		newFileName := path.Join(*outdir, a.FileName)
		err = ioutil.WriteFile(newFileName, a.Content, 0644)
		if err != nil {
			fmt.Printf("Error writing file %q: %v\n", newFileName, err)
			break
		}
	}
	fmt.Fprintln(os.Stderr, " Done!")
}
