// Package main outputs a markdown formatted document describing the provided email
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/cmd"
)

type dumper struct {
	errOut, stdOut io.Writer
	exit           exitFunc
}
type exitFunc func(int)

func newDefaultDumper() *dumper {
	return &dumper{
		errOut: os.Stderr,
		stdOut: os.Stdout,
		exit:   os.Exit,
	}
}

func main() {
	d := newDefaultDumper()
	d.exit(d.dump(os.Args))
}

func (d *dumper) dump(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(d.errOut, "Missing filename argument")
		return 1
	}

	reader, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintln(d.errOut, "Failed to open file:", err)
		return 1
	}

	// basename is used as the markdown title
	basename := filepath.Base(args[1])
	e, err := enmime.ReadEnvelope(reader)
	if err != nil {
		fmt.Fprintln(d.errOut, "During enmime.ReadEnvelope:", err)
		return 1
	}

	if err = cmd.EnvelopeToMarkdown(d.stdOut, e, basename); err != nil {
		fmt.Fprintln(d.errOut, err)
		return 1
	}
	return 0
}
