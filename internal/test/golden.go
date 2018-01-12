package test

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "Update .golden files")

// DiffLines does a line by line comparison of got and want, reporting up to five
// differences before giving up
func DiffLines(t *testing.T, got []byte, want []byte) {
	t.Helper()
	gbuf := bytes.NewBuffer(got)
	wbuf := bytes.NewBuffer(want)
	diffs := 0
	for line := 1; diffs < 5; line++ {
		g, gerr := gbuf.ReadString('\n')
		w, werr := wbuf.ReadString('\n')
		// fmt.Printf("g: %q, err: %v\n", g, gerr)
		// fmt.Printf("w: %q, err: %v\n", w, werr)
		if g != w {
			// We compare before EOF test in case the final line has no \n
			diffs++
			t.Errorf("Line %v differed\n got: %q\nwant: %q", line, g, w)
		}
		if gerr == io.EOF && werr == io.EOF {
			return
		}
		if gerr != nil {
			t.Fatalf("Error on got: %s", gerr)
			return
		}
		if werr != nil {
			t.Fatalf("Error on want: %s", werr)
			return
		}
	}
	t.Fatalf("Reached maximum of %v differences", diffs)
}

func DiffGolden(t *testing.T, got []byte, path ...string) {
	t.Helper()
	pathstr := filepath.Join(path...)
	f, err := os.Open(pathstr)
	if err != nil {
		t.Fatal(err)
	}
	golden, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, golden) {
		if *update {
			// Update golden file
			if err := ioutil.WriteFile(pathstr, got, 0666); err != nil {
				t.Fatal(err)
			}
		} else {
			// Fail test with differences
			DiffLines(t, got, golden)
		}
	}
}
