package enmime_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/jhillyerd/enmime/v2"
	"github.com/jhillyerd/enmime/v2/internal/test"
	"github.com/stretchr/testify/assert"
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

func TestEncodePartMessageType(t *testing.T) {
	p := enmime.NewPart("MeSSaGe/rfc822")
	p.Boundary = "enmime-abcdefg0123456789"

	c := make([]byte, 2000)
	for i := range c {
		c[i] = byte(i % 256)
	}
	p.Content = c

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	// Expects the binary content to be written without any additional transfer encoding.
	test.DiffGolden(t, b.Bytes(), "testdata", "encode", "part-message-rfc822.golden")
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

func TestEncodePartWithForceQuotedPrintableCte(t *testing.T) {
	nonASCIIcontent := bytes.Repeat([]byte{byte(0x10)}, 10)
	p := enmime.NewPart("text/plain").WithEncoder(enmime.NewEncoder(enmime.ForceQuotedPrintableCte(true)))
	p.Content = nonASCIIcontent
	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	// Verify output is QP encoded.
	assert.Equal(t, "quoted-printable", p.Header.Get("Content-Transfer-Encoding"))
	assert.Contains(t, b.String(), "=10=10=10")
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

// TestEncodePartForcedCTE7Bit verifies that a non-text part with 7bit-safe content and forced "7bit" CTE
// is not base64-encoded and the header reads Content-Transfer-Encoding: 7bit.
func TestEncodePartForcedCTE7Bit(t *testing.T) {
	p := enmime.NewPart("application/pgp-encrypted")
	p.ContentTransferEncoding = "7bit"
	p.Content = []byte("Version: 1")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "7bit", p.Header.Get("Content-Transfer-Encoding"))
	assert.Contains(t, b.String(), "Version: 1")
	assert.NotContains(t, b.String(), "base64")
}

// TestEncodePartForcedCTEQuotedPrintable verifies QP encoding is applied when forced.
func TestEncodePartForcedCTEQuotedPrintable(t *testing.T) {
	p := enmime.NewPart("application/pgp-signature")
	p.ContentTransferEncoding = "quoted-printable"
	p.Content = []byte("-----BEGIN PGP SIGNATURE-----\r\nVersion: GnuPG\r\n\r\n=AAAA\r\n-----END PGP SIGNATURE-----")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "quoted-printable", p.Header.Get("Content-Transfer-Encoding"))
	assert.Contains(t, b.String(), "-----BEGIN PGP")
}

// TestEncodePartForcedCTEBase64 verifies base64 is applied when explicitly forced.
func TestEncodePartForcedCTEBase64(t *testing.T) {
	p := enmime.NewPart("application/zip")
	p.ContentTransferEncoding = "base64"
	p.Content = []byte("ZIPZIPZIP")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "base64", p.Header.Get("Content-Transfer-Encoding"))
}

// TestEncodePartForcedCTE8Bit verifies 8bit header is set and content is written verbatim.
func TestEncodePartForcedCTE8Bit(t *testing.T) {
	p := enmime.NewPart("application/octet-stream")
	p.ContentTransferEncoding = "8bit"
	p.Content = []byte("some binary data")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "8bit", p.Header.Get("Content-Transfer-Encoding"))
	assert.Contains(t, b.String(), "some binary data")
}

// TestEncodePartForcedCTEEmpty verifies empty ContentTransferEncoding preserves default behavior (base64 for non-text).
func TestEncodePartForcedCTEEmpty(t *testing.T) {
	p := enmime.NewPart("application/octet-stream")
	p.ContentTransferEncoding = ""
	p.Content = []byte("data")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "base64", p.Header.Get("Content-Transfer-Encoding"))
}

// TestEncodePartForcedCTEUnrecognised verifies unrecognised values fall back to automatic detection.
func TestEncodePartForcedCTEUnrecognised(t *testing.T) {
	p := enmime.NewPart("application/octet-stream")
	p.ContentTransferEncoding = "unknown-encoding"
	p.Content = []byte("data")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	// Non-text part should default to base64.
	assert.Equal(t, "base64", p.Header.Get("Content-Transfer-Encoding"))
}

// TestEncodePartForcedCTECaseInsensitive verifies case-insensitive matching.
func TestEncodePartForcedCTECaseInsensitive(t *testing.T) {
	for _, val := range []string{"7Bit", "7BIT", "7bit"} {
		p := enmime.NewPart("application/pgp-encrypted")
		p.ContentTransferEncoding = val
		p.Content = []byte("Version: 1")

		b := &bytes.Buffer{}
		err := p.Encode(b)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "7bit", p.Header.Get("Content-Transfer-Encoding"),
			"expected 7bit for input %q", val)
	}

	for _, val := range []string{"BASE64", "Base64", "base64"} {
		p := enmime.NewPart("text/plain")
		p.ContentTransferEncoding = val
		p.Content = []byte("hello")

		b := &bytes.Buffer{}
		err := p.Encode(b)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "base64", p.Header.Get("Content-Transfer-Encoding"),
			"expected base64 for input %q", val)
	}
}

// TestEncodePartForcedCTETextPartBase64 verifies that forcing base64 on a text part
// overrides the auto-detection that would normally pick 7bit or QP.
func TestEncodePartForcedCTETextPartBase64(t *testing.T) {
	p := enmime.NewPart("text/plain")
	p.ContentTransferEncoding = "base64"
	p.Content = []byte("Hello, this is plain ASCII text that would normally be 7bit.")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "base64", p.Header.Get("Content-Transfer-Encoding"))
	// Should not contain the plaintext directly
	assert.NotContains(t, b.String(), "Hello, this is plain ASCII")
}

// TestEncodePartForcedCTEWithContentReader verifies that ContentReader + forced 7bit CTE
// writes content verbatim instead of base64-encoding it.
func TestEncodePartForcedCTEWithContentReader(t *testing.T) {
	content := []byte("Version: 1")
	p := enmime.NewPart("application/pgp-encrypted")
	p.ContentTransferEncoding = "7bit"
	p.ContentReader = bytes.NewReader(content)

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "7bit", p.Header.Get("Content-Transfer-Encoding"))
	assert.Contains(t, b.String(), "Version: 1")
	assert.Nil(t, p.ContentReader, "ContentReader should have been consumed")
}

// TestEncodePartForcedCTEPrecedenceOverEncoderOption verifies that ContentTransferEncoding on the part
// takes precedence over the encoder-level ForceQuotedPrintableCte option.
func TestEncodePartForcedCTEPrecedenceOverEncoderOption(t *testing.T) {
	p := enmime.NewPart("application/octet-stream").WithEncoder(
		enmime.NewEncoder(enmime.ForceQuotedPrintableCte(true)),
	)
	p.ContentTransferEncoding = "7bit"
	p.Content = []byte("plain ASCII data")

	b := &bytes.Buffer{}
	err := p.Encode(b)
	if err != nil {
		t.Fatal(err)
	}

	// Part-level override should win over encoder option.
	assert.Equal(t, "7bit", p.Header.Get("Content-Transfer-Encoding"))
	assert.Contains(t, b.String(), "plain ASCII data")
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
