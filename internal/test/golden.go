package test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "Update .golden files")

type section struct {
	ctype byte
	s     []string
}

// Inspired by https://github.com/paulgb/simplediff
func diff(before, after []string) []section {
	beforeMap := make(map[string][]int)
	for i, s := range before {
		beforeMap[s] = append(beforeMap[s], i)
	}
	overlap := make([]int, len(before))
	// Track start/len of largest overlapping match in old/new
	var startBefore, startAfter, subLen int
	for iafter, s := range after {
		o := make([]int, len(before))
		for _, ibefore := range beforeMap[s] {
			idx := 1
			if ibefore > 0 && overlap[ibefore-1] > 0 {
				idx = overlap[ibefore-1] + 1
			}
			o[ibefore] = idx
			if idx > subLen {
				// largest substring so far, store indices
				subLen = o[ibefore]
				startBefore = ibefore - subLen + 1
				startAfter = iafter - subLen + 1
			}
		}
		overlap = o
	}
	if subLen == 0 {
		// No common substring, issue - and +
		r := make([]section, 0)
		if len(before) > 0 {
			r = append(r, section{'-', before})
		}
		if len(after) > 0 {
			r = append(r, section{'+', after})
		}
		return r
	}
	// common substring unchanged, recurse on before/after substring
	r := diff(before[0:startBefore], after[0:startAfter])
	r = append(r, section{' ', after[startAfter : startAfter+subLen]})
	r = append(r, diff(before[startBefore+subLen:], after[startAfter+subLen:])...)
	return r
}

// DiffStrings does a entry by entry comparison of got and want.
func DiffStrings(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) == 0 && len(want) == 0 {
		return
	}
	sections := diff(want, got)
	if len(sections) == 1 && sections[0].ctype == ' ' {
		// Equal
		return
	}
	t.Error("diff -want +got:")
	for _, s := range sections {
		if s.ctype == ' ' && len(s.s) > 5 {
			// Omit excess unchanged lines
			for i := 0; i < 2; i++ {
				t.Logf("|%c%s", s.ctype, s.s[i])
			}
			t.Log("...")
			for i := len(s.s) - 2; i < len(s.s); i++ {
				t.Logf("|%c%s", s.ctype, s.s[i])
			}
			continue
		}
		for _, l := range s.s {
			t.Logf("|%c%s", s.ctype, l)
		}
	}
}

// DiffLines does a line by line comparison of got and want.
func DiffLines(t *testing.T, got []byte, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		b := bytes.NewBufferString("diff -want +got:\n")
		glines := strings.Split(string(got), "\n")
		wlines := strings.Split(string(want), "\n")
		sections := diff(wlines, glines)
		for _, s := range sections {
			if s.ctype == ' ' && len(s.s) > 5 {
				// Omit excess unchanged lines
				for i := 0; i < 2; i++ {
					fmt.Fprintf(b, "|%c%s\n", s.ctype, s.s[i])
				}
				b.WriteString("...\n")
				for i := len(s.s) - 2; i < len(s.s); i++ {
					fmt.Fprintf(b, "|%c%s\n", s.ctype, s.s[i])
				}
				continue
			}
			for _, l := range s.s {
				fmt.Fprintf(b, "|%c%s\n", s.ctype, l)
			}
		}
		t.Error(b.String())
	}
}

// DiffGolden does a line by comparison of got to the golden file specified by path. If the update
// flag is true, differing golden files will be updated with lines in got.
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
			t.Errorf("Test output did not match %s\nTo update golden file, run: go test -update", pathstr)
			// Fail test with differences
			DiffLines(t, got, golden)
		}
	}
}
