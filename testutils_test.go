package enmime

import (
	"io"
	"os"
	"path/filepath"
)

// openTestData is a utility function to open a file in testdata for reading, it will panic if there
// is an error.
func openTestData(subdir, filename string) io.Reader {
	// Open test part for parsing
	raw, err := os.Open(filepath.Join("testdata", subdir, filename))
	if err != nil {
		// err already contains full path to file
		panic(err)
	}
	return raw
}
