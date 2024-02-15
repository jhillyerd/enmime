package enmime_test

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
	"time"

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
	p := enmime.NewPart("text/plain")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-header-only.golden")
}

func TestEncodePartHeaderOnlyDefaultTransferEncoding(t *testing.T) {
	p := enmime.NewPart("text/plain")
	p.Header.Add("X-Empty-Header", "")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-header-only-default-encoding.golden")
}

func TestEncodePartDefaultHeaders(t *testing.T) {
	p := enmime.NewPart("application/zip")
	p.Boundary = "enmime-abcdefg0123456789"
	p.Charset = "binary"
	p.ContentID = "mycontentid"
	p.ContentTypeParams["param1"] = "myparameter1"
	p.ContentTypeParams["param2"] = "myparameter2"
	p.Disposition = "attachment"
	p.FileName = "stuff.zip"
	p.FileModDate, _ = time.Parse(time.RFC822, "01 Feb 03 04:05 GMT")
	p.Content = []byte("ZIPZIPZIP")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-default-headers.golden")
}

func TestEncodePartQuotedHeaders(t *testing.T) {
	p := enmime.NewPart("application/zip")
	p.Boundary = "enmime-abcdefg0123456789"
	p.Charset = "binary"
	p.ContentID = "mycontentid"
	p.ContentTypeParams["param1"] = "myparameter1"
	p.ContentTypeParams["param2"] = "myparameter2"
	p.Disposition = "attachment"
	p.FileName = `árvíztűrő "x" tükörfúrógép.zip`
	p.FileModDate, _ = time.Parse(time.RFC822, "01 Feb 03 04:05 GMT")
	p.Content = []byte("ZIPZIPZIP")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-quoted-headers.golden")
}

func TestEncodePartQuotedPrintableHeaders(t *testing.T) {
	p := enmime.NewPart("application/zip")
	p.Boundary = "enmime-abcdefg0123456789"
	p.Charset = "binary"
	p.ContentID = "mycontentid"
	p.ContentTypeParams["param1"] = "myparameter1"
	p.ContentTypeParams["param2"] = "myparameter2"
	p.Disposition = "attachment"
	p.FileName = `árvíztűrő "x" tükörfúrógép.zip`
	p.FileModDate, _ = time.Parse(time.RFC822, "01 Feb 03 04:05 GMT")
	p.Header.Add("X-QP-Header", "Just enough to need qp ☆")
	p.Content = []byte("ZIPZIPZIP")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-quoted-printable-headers.golden")
}

// oneByOneReader implements io.Reader over a byte slice, returns a single byte on every Read request.
// This object is used to validate that partial reads (=read calls that return n<len(p)) are handled correctly.
type oneByOneReader struct {
	content []byte
	pos     int
}

func (r *oneByOneReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.pos >= len(r.content) {
		return 0, io.EOF
	}
	p[0] = r.content[r.pos]
	r.pos++
	return 1, nil
}

func TestEncodePartContentReader(t *testing.T) {
	contentLengths := []int{
		0, 1, 2, 3, 4, // empty / nearly empty
		55, 56, 57, 58, 59, 60, // lengths close to the length of a single line (57)
		7294, 7295, 7296, 7297, 7298, // lengths close to the length of a single chunk (7296)
	}

	for _, oneByOne := range []bool{false, true} {
		for _, contentLength := range contentLengths {
			// create a part with random content
			p := enmime.NewPart("application/zip")
			p.Boundary = "enmime-abcdefg0123456789"
			p.Charset = "binary"
			p.ContentID = "mycontentid"
			p.ContentTypeParams["param1"] = "myparameter1"
			p.ContentTypeParams["param2"] = "myparameter2"
			p.Disposition = "attachment"
			p.FileName = "stuff.zip"
			p.FileModDate, _ = time.Parse(time.RFC822, "01 Feb 03 04:05 GMT")

			p.Content = make([]byte, contentLength)
			_, err := rand.Read(p.Content)
			if err != nil {
				t.Fatal(err)
			}

			// encode the part using `Content` byte slice stored in the Part
			b1 := &bytes.Buffer{}
			err = p.Encode(b1)
			if err != nil {
				t.Fatal(err)
			}

			// encode the part using io.reader
			if oneByOne {
				p.ContentReader = &oneByOneReader{content: p.Content}
			} else {
				p.ContentReader = bytes.NewReader(p.Content)
			}
			p.Content = nil

			b2 := &bytes.Buffer{}
			err = p.Encode(b2)
			if err != nil {
				t.Fatal(err)
			}

			// compare the results
			if !bytes.Equal(b1.Bytes(), b2.Bytes()) {
				t.Errorf("[]byte encode and io.Reader encode produced different results for length %d", contentLength)
			}
		}
	}
}

func TestEncodePartBinaryHeader(t *testing.T) {
	p := enmime.NewPart("text/plain")
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

func TestEncodePartContentOnlyQP(t *testing.T) {
	p := &enmime.Part{}
	p.Content = []byte("☆ No header, only content.")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-content-only-qp.golden")
}

func TestEncodePartPlain(t *testing.T) {
	p := enmime.NewPart("text/plain")
	p.Content = []byte("This is a test of a plain text part.\r\n\r\nAnother line.\r\n")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-plain.golden")
}

func TestEncodePartWithChildren(t *testing.T) {
	p := enmime.NewPart("multipart/alternative")
	p.Boundary = "enmime-1234567890-parent"
	p.Content = []byte("Bro, do you even MIME?")
	root := p

	p = enmime.NewPart("text/html")
	p.Content = []byte("<div>HTML part</div>")
	root.FirstChild = p

	p = enmime.NewPart("text/plain")
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
	p := enmime.NewPart("multipart/alternative")
	p.Boundary = "enmime-1234567890-parent"
	root := p

	p = enmime.NewPart("text/html")
	p.Content = []byte("<div>HTML part</div>")
	root.FirstChild = p

	p = enmime.NewPart("text/plain")
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
	p := enmime.NewPart("text/plain")
	p.Content = []byte("¡Hola, señor! Welcome to MIME")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-quoted-content.golden")
}

func TestEncodePartWithExistingEncodingHeader(t *testing.T) {
	p := enmime.NewPart("text/plain")
	p.Header.Add("Content-Transfer-Encoding", "quoted-printable")
	p.Content = []byte("Hello=")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-quotable-content.golden")
}

func TestEncodePartContentBinary(t *testing.T) {
	c := make([]byte, 2000)
	for i := range c {
		c[i] = byte(i % 256)
	}
	p := enmime.NewPart("image/jpeg")
	p.Content = c

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-bin-content.golden")
}

func TestEncodeFileModDate(t *testing.T) {
	p := enmime.NewPart("text/plain")
	p.Content = []byte("¡Hola, señor! Welcome to MIME")
	p.Disposition = "inline"
	p.FileModDate, _ = time.Parse(time.RFC822, "01 Feb 03 04:05 GMT")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-file-mod-date.golden")
}

func TestEncodePartContentNonAsciiText(t *testing.T) {
	p := enmime.NewPart("text/plain")

	threshold := 20

	cases := []int{
		threshold - 1,
		threshold,
		threshold + 1,
	}

	for _, numNonASCII := range cases {
		nonASCII := bytes.Repeat([]byte{byte(0x10)}, numNonASCII)
		ascii := bytes.Repeat([]byte{0x41}, 100-numNonASCII)
		nonASCII = append(nonASCII, ascii...)

		p.Content = nonASCII
		b := &bytes.Buffer{}
		err := p.Encode(b)
		if err != nil {
			t.Fatal(err)
		}

		if numNonASCII < threshold {
			test.DiffStrings(t, []string{p.Header.Get("Content-Transfer-Encoding")}, []string{"quoted-printable"})
		} else {
			test.DiffStrings(t, []string{p.Header.Get("Content-Transfer-Encoding")}, []string{"base64"})
		}
	}
}

// TestParseRawContentHTMLOptionTrue tests the RawContent Parser option with an HTML part only if the content isn't changed.
func TestParseRawContentHTMLOptionTrue(t *testing.T) {
	r := test.OpenTestData("encode", "parser-raw-content-html-option.raw")
	p := enmime.NewParser(enmime.RawContent(true))
	e, err := p.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	b := &bytes.Buffer{}
	if err := e.Root.Encode(b); err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "parser-raw-content-html-option-true.raw.golden")
}

// TestParseRawContentHTMLOptionFalse tests without the RawContent Parser option and an HTML part only if the content is normally parsed.
func TestParseRawContentHTMLOptionFalse(t *testing.T) {
	r := test.OpenTestData("encode", "parser-raw-content-html-option.raw")
	p := enmime.NewParser()
	e, err := p.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	b := &bytes.Buffer{}
	if err := e.Root.Encode(b); err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "parser-raw-content-html-option-false.raw.golden")
}

// TestParseRawContentTextOptionTrue tests the RawContent Parser option with a TEXT part only if the content isn't changed.
// This test uses Japanese characters to also ensure that the charset isn't altered.
func TestParseRawContentTextOptionTrue(t *testing.T) {
	r := test.OpenTestData("encode", "parser-raw-content-text-option.raw")
	p := enmime.NewParser(enmime.RawContent(true))
	e, err := p.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	b := &bytes.Buffer{}
	if err := e.Root.Encode(b); err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "parser-raw-content-text-option-true.raw.golden")
}

// TestParseRawContentTextOptionFalse tests without the RawContent Parser option and a TEXT part only if the content is normally parsed.
func TestParseRawContentTextOptionFalse(t *testing.T) {
	r := test.OpenTestData("encode", "parser-raw-content-text-option.raw")
	p := enmime.NewParser()
	e, err := p.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	b := &bytes.Buffer{}
	if err := e.Root.Encode(b); err != nil {
		t.Fatal(err)
	}
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "parser-raw-content-text-option-false.raw.golden")
}

// TestRawContentUTF8Headers verifies plain-text headers are unmodified with the rawContent parser option.
func TestRawContentUTF8Headers(t *testing.T) {
	r := test.OpenTestData("encode", "utf8-to.raw.golden")
	p := enmime.NewParser(enmime.RawContent(true))
	e, err := p.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	b := &bytes.Buffer{}
	if err := e.Root.Encode(b); err != nil {
		t.Fatal(err)
	}

	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "utf8-to.raw.golden")
}
