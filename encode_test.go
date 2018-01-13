package enmime_test

import (
	"bytes"
	"testing"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/internal/test"
)

func TestEncodePartEmpty(t *testing.T) {
	p := &enmime.Part{}

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-empty.golden")
}

func TestEncodePartHeaderOnly(t *testing.T) {
	p := enmime.NewPart(nil, "text/plain")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-header-only.golden")
}

func TestEncodePartContentOnly(t *testing.T) {
	p := &enmime.Part{}
	p.Content = []byte("No header, only content.")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-content-only.golden")
}

func TestEncodePartPlain(t *testing.T) {
	p := enmime.NewPart(nil, "text/plain")
	p.Content = []byte("This is a test of a plain text part.\r\n\r\nAnother line.\r\n")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-plain.golden")
}

func TestEncodePartWithChildren(t *testing.T) {
	p := enmime.NewPart(nil, "multipart/alternative")
	p.Boundary = "enmime-1234567890-parent"
	p.Content = []byte("Do you even MIME bro?")
	root := p

	p = enmime.NewPart(root, "text/html")
	p.Content = []byte("<div>HTML part</div>")
	root.FirstChild = p

	p = enmime.NewPart(root, "text/plain")
	p.Content = []byte("Plain text part")
	root.FirstChild.NextSibling = p

	b := &bytes.Buffer{}
	err := root.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-with-children.golden")
}
