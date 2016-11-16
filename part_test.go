package enmime

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlainTextPart(t *testing.T) {
	r := openPart("textplain.raw")
	p, err := ReadParts(r)

	if err != nil {
		t.Fatal("Parsing should not have generated an error")
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	want := "7bit"
	got := p.Header.Get("Content-Transfer-Encoding")
	if got != want {
		t.Errorf("Content-Transfer-Encoding got: %q, want: %q", got, want)
	}

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "Test of text/plain section"
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}

	if p.NextSibling() != nil {
		t.Error("Root should never have a sibling")
	}
}

func TestQuotedPrintablePart(t *testing.T) {
	r := openPart("quoted-printable.raw")
	p, err := ReadParts(r)

	if err != nil {
		t.Fatal("Parsing should not have generated an error")
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	want := "quoted-printable"
	got := p.Header.Get("Content-Transfer-Encoding")
	if got != want {
		t.Errorf("Content-Transfer-Encoding got: %q, want: %q", got, want)
	}

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "Start=ABC=Finish"
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}
	if p.NextSibling() != nil {
		t.Error("Root should never have a sibling")
	}

}

func TestMultiAlternParts(t *testing.T) {
	r := openPart("multialtern.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Parsing should not have generated an error")
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	want := "multipart/alternative"
	got := p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content() should have length of 0")
	}
	if p.FirstChild() == nil {
		t.Error("Root should have a FirstChild")
	}
	if p.NextSibling() != nil {
		t.Error("Root should never have a sibling")
	}

	// Examine first child
	p = p.FirstChild()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "A text section"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() == nil {
		t.Error("First child should have a sibling")
	}

	// Examine sibling
	p = p.NextSibling()

	want = "text/html"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "An HTML section"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() != nil {
		t.Error("NextSibling() should be nil")
	}
}

func TestMultiMixedParts(t *testing.T) {
	r := openPart("multimixed.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Parsing should not have generated an error")
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	want := "multipart/mixed"
	got := p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content() should have length of 0")
	}
	if p.FirstChild() == nil {
		t.Error("Root should have a FirstChild")
	}
	if p.NextSibling() != nil {
		t.Error("Root should never have a sibling")
	}

	// Examine first child
	p = p.FirstChild()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "Section one"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() == nil {
		t.Error("First child should have a sibling")
	}

	// Examine sibling
	p = p.NextSibling()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "Section two"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() != nil {
		t.Error("NextSibling() should be nil")
	}
}

func TestMultiOtherParts(t *testing.T) {
	r := openPart("multiother.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Parsing should not have generated an error")
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	want := "multipart/x-enmime"
	got := p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content() should have length of 0")
	}
	if p.FirstChild() == nil {
		t.Error("Root should have a FirstChild")
	}
	if p.NextSibling() != nil {
		t.Error("Root should never have a sibling")
	}

	// Examine first child
	p = p.FirstChild()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "Section one"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() == nil {
		t.Error("First child should have a sibling")
	}

	// Examine sibling
	p = p.NextSibling()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "Section two"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() != nil {
		t.Error("NextSibling() should be nil")
	}
}

func TestNestedAlternParts(t *testing.T) {
	r := openPart("nestedmulti.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Parsing should not have generated an error")
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	want := "multipart/alternative"
	got := p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content() should have length of 0")
	}
	if p.FirstChild() == nil {
		t.Error("Root should have a FirstChild")
	}
	if p.NextSibling() != nil {
		t.Error("Root should never have a sibling")
	}

	// Examine first child
	p = p.FirstChild()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "A text section"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() == nil {
		t.Error("First child should have a sibling")
	}

	// Examine sibling
	p = p.NextSibling()

	want = "multipart/related"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content() should have length of 0")
	}
	if p.NextSibling() != nil {
		t.Error("NextSibling() should be nil")
	}
	if p.FirstChild() == nil {
		t.Error("Second child should have a child")
	}

	// First nested
	p = p.FirstChild()

	want = "text/html"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "An HTML section"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() == nil {
		t.Error("First nested should have a sibling")
	}

	// Second nested
	p = p.NextSibling()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "inline"
	got = p.Disposition()
	if got != want {
		t.Errorf("Disposition() got: %q, want: %q", got, want)
	}

	want = "attach.txt"
	got = p.FileName()
	if got != want {
		t.Errorf("FileName() got: %q, want: %q", got, want)
	}

	want = "An inline text attachment"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() == nil {
		t.Error("Second nested should have a sibling")
	}

	// Third nested
	p = p.NextSibling()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "inline"
	got = p.Disposition()
	if got != want {
		t.Errorf("Disposition() got: %q, want: %q", got, want)
	}

	want = "attach2.txt"
	got = p.FileName()
	if got != want {
		t.Errorf("FileName() got: %q, want: %q", got, want)
	}

	want = "Another inline text attachment"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() != nil {
		t.Error("NextSibling() should be nil")
	}
}

func TestMultiBase64Parts(t *testing.T) {
	r := openPart("multibase64.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Parsing should not have generated an error")
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	want := "multipart/mixed"
	got := p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content() should have length of 0")
	}
	if p.FirstChild() == nil {
		t.Error("Root should have a FirstChild")
	}
	if p.NextSibling() != nil {
		t.Error("Root should never have a sibling")
	}

	// Examine first child
	p = p.FirstChild()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "A text section"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() == nil {
		t.Error("First child should have a sibling")
	}

	// Examine sibling
	p = p.NextSibling()

	want = "text/html"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}
	if p.NextSibling() != nil {
		t.Error("NextSibling() should be nil")
	}
	if p.FirstChild() != nil {
		t.Error("FirstChild() should be nil")
	}

	want = "<html>"
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
}

func TestBadBoundaryTerm(t *testing.T) {
	r := openPart("badboundary.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Parsing should not have generated an error")
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	want := "multipart/alternative"
	got := p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	// Examine first child
	p = p.FirstChild()

	want = "text/plain"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}
	if p.NextSibling() == nil {
		t.Error("First child should have a sibling")
	}

	// Examine sibling
	p = p.NextSibling()

	want = "text/html"
	got = p.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "An HTML section"
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q, should contain: %q", got, want)
	}
	if p.NextSibling() != nil {
		t.Error("NextSibling() should be nil")
	}
}

func TestPartSetter(t *testing.T) {
	m := Part{
		Header: textproto.MIMEHeader{
			"Content-Type": {"testType"},
		},
	}

	want := "application/octet-stream"
	m.SetContentType(want)
	got := m.ContentType()
	if got != want {
		t.Errorf("ContentType() got: %q, want: %q", got, want)
	}

	want = "inline"
	m.SetDisposition(want)
	got = m.Disposition()
	if got != want {
		t.Errorf("Disposition() got: %q, want: %q", got, want)
	}

	want = "somefilename"
	m.SetFileName(want)
	got = m.FileName()
	if got != want {
		t.Errorf("FileName() got: %q, want: %q", got, want)
	}
}

// openPart is a test utility function to open a part as a reader
func openPart(filename string) *bufio.Reader {
	// Open test part for parsing
	raw, err := os.Open(filepath.Join("testdata", "parts", filename))
	if err != nil {
		panic(fmt.Sprintf("Failed to open test data: %v", err))
	}

	// Wrap in a buffer
	return bufio.NewReader(raw)
}
