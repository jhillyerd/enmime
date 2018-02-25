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

func TestEncodePartDefaultHeaders(t *testing.T) {
	p := enmime.NewPart(nil, "application/zip")
	p.Boundary = "enmime-abcdefg0123456789"
	p.Charset = "binary"
	p.ContentID = "mycontentid"
	p.Disposition = "attachment"
	p.FileName = "stuff.zip"
	p.Content = []byte("ZIPZIPZIP")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-default-headers.golden")
}

func TestEncodePartQuotedHeaders(t *testing.T) {
	p := enmime.NewPart(nil, "application/zip")
	p.Boundary = "enmime-abcdefg0123456789"
	p.Charset = "binary"
	p.ContentID = "mycontentid"
	p.Disposition = "attachment"
	p.FileName = `árvíztűrő "x" tükörfúrógép.zip`
	p.Content = []byte("ZIPZIPZIP")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-quoted-headers.golden")
}

func TestEncodePartBinaryHeader(t *testing.T) {
	p := enmime.NewPart(nil, "text/plain")
	p.Header.Set("Subject", "¡Hola, señor!")
	p.Header.Set("X-Data", string([]byte{
		0x3, 0x17, 0xe1, 0x7e, 0xe8, 0xeb, 0xa2, 0x96, 0x9d, 0x95, 0xa7, 0x67, 0x82, 0x9,
		0xdf, 0x8e, 0xc, 0x2c, 0x6a, 0x2b, 0x9b, 0xbe, 0x79, 0xa4, 0x69, 0xd8, 0xae, 0x86,
		0xd7, 0xab, 0xa8, 0x72, 0x52, 0x15, 0xfb, 0x80, 0x8e, 0x47, 0xe1, 0xae, 0xaa, 0x5e,
		0xa2, 0xb2, 0xc0, 0x90, 0x59, 0xe3, 0x35, 0xf8, 0x60, 0xb7, 0xb1, 0x63, 0x77, 0xd7,
		0x5f, 0x92, 0x58, 0xa8, 0x75,
	}))
	p.Content = []byte("This is a test of a plain text part.\r\n\r\nAnother line.\r\n")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-bin-header.golden")
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

func TestEncodePartNoContentWithChildren(t *testing.T) {
	p := enmime.NewPart(nil, "multipart/alternative")
	p.Boundary = "enmime-1234567890-parent"
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
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "nocontent-with-children.golden")
}

func TestEncodePartContentQuotable(t *testing.T) {
	p := enmime.NewPart(nil, "text/plain")
	p.Content = []byte("¡Hola, señor! Welcome to MIME")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-quoted-content.golden")
}

func TestEncodePartContentBinary(t *testing.T) {
	c := make([]byte, 2000)
	for i := range c {
		c[i] = byte(i % 256)
	}
	p := enmime.NewPart(nil, "image/jpeg")
	p.Content = c

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-bin-content.golden")
}
