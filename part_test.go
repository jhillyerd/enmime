package enmime_test

import (
	"testing"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/internal/test"
)

func TestPlainTextPart(t *testing.T) {
	var want, got string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "textplain.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	want = "7bit"
	got = p.Header.Get("Content-Transfer-Encoding")
	if got != want {
		t.Errorf("Content-Transfer-Encoding got: %q, want: %q", got, want)
	}

	want = "Test of text/plain section"
	test.ContentContainsString(t, p.Content, want)
}

func TestQuotedPrintablePart(t *testing.T) {
	var want, got string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "quoted-printable.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	want = "quoted-printable"
	got = p.Header.Get("Content-Transfer-Encoding")
	if got != want {
		t.Errorf("Content-Transfer-Encoding got: %q, want: %q", got, want)
	}

	want = "Start=ABC=Finish"
	test.ContentEqualsString(t, p.Content, want)
}

func TestQuotedPrintableInvalidPart(t *testing.T) {
	var want, got string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "quoted-printable-invalid.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		ContentType: "text/plain",
		Charset:     "utf-8",
		Disposition: "inline",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	want = "quoted-printable"
	got = p.Header.Get("Content-Transfer-Encoding")
	if got != want {
		t.Errorf("Content-Transfer-Encoding got: %q, want: %q", got, want)
	}

	want = "Stuffs’s Weekly Summary"
	test.ContentContainsString(t, p.Content, want)
}

func TestMultiAlternParts(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "multialtern.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		FirstChild:  test.PartExists,
		ContentType: "multipart/alternative",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	test.ComparePart(t, p, wantp)

	want = "A text section"
	test.ContentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	test.ComparePart(t, p, wantp)

	want = "An HTML section"
	test.ContentContainsString(t, p.Content, want)
}

// TestRootMissingContentType expects a default content type to be set for the root if not specified
func TestRootMissingContentType(t *testing.T) {
	var want string
	r := test.OpenTestData("parts", "missing-ctype-root.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}
	want = "text/plain"
	if p.ContentType != want {
		t.Errorf("Content-Type got: %q, want: %q", p.ContentType, want)
	}
	want = "us-ascii"
	if p.Charset != want {
		t.Errorf("Charset got: %q, want: %q", p.Charset, want)
	}
}

func TestPartMissingContentType(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "missing-ctype.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		FirstChild:  test.PartExists,
		ContentType: "multipart/alternative",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		// No ContentType
		PartID: "1",
	}
	test.ComparePart(t, p, wantp)

	want = "A text section"
	test.ContentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	test.ComparePart(t, p, wantp)

	want = "An HTML section"
	test.ContentContainsString(t, p.Content, want)
}

func TestPartEmptyHeader(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "empty-header.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		FirstChild:  test.PartExists,
		ContentType: "multipart/alternative",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild

	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		// No ContentType
		PartID: "1",
	}
	test.ComparePart(t, p, wantp)

	want = "A text section"
	test.ContentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	test.ComparePart(t, p, wantp)

	want = "An HTML section"
	test.ContentContainsString(t, p.Content, want)
}

func TestPartHeaders(t *testing.T) {
	r := test.OpenTestData("parts", "header-only.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	want := "text/html"
	if p.ContentType != want {
		t.Errorf("ContentType %q, want %q", p.ContentType, want)
	}
	want = "file.html"
	if p.FileName != want {
		t.Errorf("FileName %q, want %q", p.FileName, want)
	}
	want = "utf-8"
	if p.Charset != want {
		t.Errorf("Charset %q, want %q", p.Charset, want)
	}
	want = "inline"
	if p.Disposition != want {
		t.Errorf("Disposition %q, want %q", p.Disposition, want)
	}
	want = "part123456@inbucket.org"
	if p.ContentID != want {
		t.Errorf("ContentID %q, want %q", p.ContentID, want)
	}
}

func TestMultiMixedParts(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "multimixed.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		FirstChild:  test.PartExists,
		ContentType: "multipart/mixed",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	test.ComparePart(t, p, wantp)

	want = "Section one"
	test.ContentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	test.ComparePart(t, p, wantp)

	want = "Section two"
	test.ContentContainsString(t, p.Content, want)
}

func TestMultiOtherParts(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "multiother.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		FirstChild:  test.PartExists,
		ContentType: "multipart/x-enmime",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	test.ComparePart(t, p, wantp)

	want = "Section one"
	test.ContentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	test.ComparePart(t, p, wantp)

	want = "Section two"
	test.ContentContainsString(t, p.Content, want)
}

func TestNestedAlternParts(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "nestedmulti.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		ContentType: "multipart/alternative",
		FirstChild:  test.PartExists,
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	test.ComparePart(t, p, wantp)

	want = "A text section"
	test.ContentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		FirstChild:  test.PartExists,
		ContentType: "multipart/related",
		PartID:      "2.0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// First nested
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2.1",
	}
	test.ComparePart(t, p, wantp)

	want = "An HTML section"
	test.ContentContainsString(t, p.Content, want)

	// Second nested
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Disposition: "inline",
		FileName:    "attach.txt",
		PartID:      "2.2",
	}
	test.ComparePart(t, p, wantp)

	want = "An inline text attachment"
	test.ContentContainsString(t, p.Content, want)

	// Third nested
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/plain",
		Disposition: "inline",
		FileName:    "attach2.txt",
		PartID:      "2.3",
	}
	test.ComparePart(t, p, wantp)

	want = "Another inline text attachment"
	test.ContentContainsString(t, p.Content, want)
}

func TestPartSimilarBoundary(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "similar-boundary.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		ContentType: "multipart/mixed",
		FirstChild:  test.PartExists,
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	test.ComparePart(t, p, wantp)

	want = "Section one"
	test.ContentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		FirstChild:  test.PartExists,
		ContentType: "multipart/alternative",
		PartID:      "2.0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// First nested
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "2.1",
	}
	test.ComparePart(t, p, wantp)

	want = "A text section"
	test.ContentContainsString(t, p.Content, want)

	// Second nested
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2.2",
	}
	test.ComparePart(t, p, wantp)

	want = "An HTML section"
	test.ContentContainsString(t, p.Content, want)
}

// Check we don't UTF-8 decode attachments
func TestBinaryDecode(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "bin-attach.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		FirstChild:  test.PartExists,
		ContentType: "multipart/mixed",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	test.ComparePart(t, p, wantp)

	want = "A text section"
	test.ContentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "application/octet-stream",
		Charset:     "us-ascii",
		Disposition: "attachment",
		FileName:    "test.bin",
		PartID:      "2",
	}
	test.ComparePart(t, p, wantp)

	wantBytes := []byte{
		0x50, 0x4B, 0x03, 0x04, 0x14, 0x00, 0x08, 0x00,
		0x08, 0x00, 0xC2, 0x02, 0x29, 0x4A, 0x00, 0x00}
	test.ContentEqualsBytes(t, p.Content, wantBytes)
}

func TestMultiBase64Parts(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "multibase64.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		FirstChild:  test.PartExists,
		ContentType: "multipart/mixed",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	test.ContentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	test.ComparePart(t, p, wantp)

	want = "A text section"
	test.ContentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/html",
		Disposition: "attachment",
		FileName:    "test.html",
		PartID:      "2",
	}
	test.ComparePart(t, p, wantp)

	want = "<html>"
	test.ContentContainsString(t, p.Content, want)
}

func TestBadBoundaryTerm(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "badboundary.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &enmime.Part{
		FirstChild:  test.PartExists,
		ContentType: "multipart/alternative",
		PartID:      "0",
	}
	test.ComparePart(t, p, wantp)

	// Examine first child
	p = p.FirstChild
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	test.ComparePart(t, p, wantp)

	// Examine sibling
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	test.ComparePart(t, p, wantp)

	want = "An HTML section"
	test.ContentContainsString(t, p.Content, want)
}
