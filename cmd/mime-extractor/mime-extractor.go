// Package main extracts attachments from the provided email
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jhillyerd/enmime/v2"
	"github.com/jhillyerd/enmime/v2/cmd"
)

var (
	mimefile = flag.String("f", "", "mime(eml) file")
	outdir   = flag.String("o", "", "output dir")
)

type extractor struct {
	errOut, stdOut io.Writer
	exit           exitFunc
	wd             workingDir
	fileWrite      attachmentWriter
}

type exitFunc func(int)
type workingDir func() (string, error)
type attachmentWriter func(string, []byte, os.FileMode) error

func newDefaultExtractor() *extractor {
	return &extractor{
		errOut:    os.Stderr,
		stdOut:    os.Stdout,
		exit:      os.Exit,
		wd:        os.Getwd,
		fileWrite: os.WriteFile,
	}
}

func main() {
	flag.Parse()
	ex := newDefaultExtractor()
	ex.exit(ex.extract(*mimefile, *outdir))
}

func (ex *extractor) extract(file, dir string) int {
	if file == "" {
		fmt.Fprintln(ex.errOut, "Missing filename argument")
		flag.Usage()
		return 1
	}
	if dir == "" {
		dir, _ = ex.wd()
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Fprintf(ex.errOut, "Mkdir %s failed.", dir)
		return 2
	}

	reader, err := os.Open(file)
	if err != nil {
		fmt.Fprintln(ex.errOut, "Failed to open file:", err)
		return 1
	}

	// basename is used as the markdown title
	basename := filepath.Base(file)
	e, err := enmime.ReadEnvelope(reader)
	if err != nil {
		fmt.Fprintln(ex.errOut, "During enmime.ReadEnvelope:", err)
		return 1
	}

	if err = cmd.EnvelopeToMarkdown(ex.stdOut, e, basename); err != nil {
		fmt.Fprintln(ex.errOut, err)
		return 1
	}

	// Write errOut attachments
	fmt.Fprintf(ex.errOut, "\nExtracting attachments into %s...", dir)
	for _, a := range e.Attachments {
		newFileName := filepath.Join(dir, a.FileName)
		err = ex.fileWrite(newFileName, a.Content, 0644)
		if err != nil {
			fmt.Fprintf(ex.stdOut, "Error writing file %q: %v\n", newFileName, err)
			break
		}
	}
	fmt.Fprintln(ex.errOut, " Done!")
	return 0
}
