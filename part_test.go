package enmime

import (
	"testing"
)

func TestPlainTextPart(t *testing.T) {
	var want, got string
	var wantp *Part
	r := openTestData("parts", "textplain.raw")
	p, err := ReadParts(r)
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		ContentType: "text/plain",
		Charset:     "us-ascii",
	}
	comparePart(t, p, wantp)

	want = "7bit"
	got = p.Header.Get("Content-Transfer-Encoding")
	if got != want {
		t.Errorf("Content-Transfer-Encoding got: %q, want: %q", got, want)
	}

	want = "Test of text/plain section"
	contentContainsString(t, p.Content, want)
}

func TestQuotedPrintablePart(t *testing.T) {
	var want, got string
	var wantp *Part
	r := openTestData("parts", "quoted-printable.raw")
	p, err := ReadParts(r)
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		ContentType: "text/plain",
		Charset:     "us-ascii",
	}
	comparePart(t, p, wantp)

	want = "quoted-printable"
	got = p.Header.Get("Content-Transfer-Encoding")
	if got != want {
		t.Errorf("Content-Transfer-Encoding got: %q, want: %q", got, want)
	}

	want = "Start=ABC=Finish"
	contentEqualsString(t, p.Content, want)
}

func TestQuotedPrintableInvalidPart(t *testing.T) {
	var want, got string
	var wantp *Part
	r := openTestData("parts", "quoted-printable-invalid.raw")
	p, err := ReadParts(r)
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		ContentType: "text/plain",
		Charset:     "utf-8",
	}
	comparePart(t, p, wantp)

	want = "quoted-printable"
	got = p.Header.Get("Content-Transfer-Encoding")
	if got != want {
		t.Errorf("Content-Transfer-Encoding got: %q, want: %q", got, want)
	}

	want = "Stuffs’s Weekly Summary"
	contentContainsString(t, p.Content, want)
}

func TestMultiAlternParts(t *testing.T) {
	var want string
	var wantp *Part
	r := openTestData("parts", "multialtern.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		FirstChild:  partExists,
		ContentType: "multipart/alternative",
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	comparePart(t, p, wantp)

	want = "A text section"
	contentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	comparePart(t, p, wantp)

	want = "An HTML section"
	contentContainsString(t, p.Content, want)
}

// TestRootMissingContentType expects a default content type to be set for the root if not specified
func TestRootMissingContentType(t *testing.T) {
	var want string
	r := openTestData("parts", "missing-ctype-root.raw")
	p, err := ReadParts(r)

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
	var wantp *Part
	r := openTestData("parts", "missing-ctype.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		FirstChild:  partExists,
		ContentType: "multipart/alternative",
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		// No ContentType
		PartID: "1",
	}
	comparePart(t, p, wantp)

	want = "A text section"
	contentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	comparePart(t, p, wantp)

	want = "An HTML section"
	contentContainsString(t, p.Content, want)
}

func TestPartEmptyHeader(t *testing.T) {
	var want string
	var wantp *Part
	r := openTestData("parts", "empty-header.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		FirstChild:  partExists,
		ContentType: "multipart/alternative",
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild

	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		// No ContentType
		PartID: "1",
	}
	comparePart(t, p, wantp)

	want = "A text section"
	contentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	comparePart(t, p, wantp)

	want = "An HTML section"
	contentContainsString(t, p.Content, want)
}

func TestMultiMixedParts(t *testing.T) {
	var want string
	var wantp *Part
	r := openTestData("parts", "multimixed.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		FirstChild:  partExists,
		ContentType: "multipart/mixed",
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	comparePart(t, p, wantp)

	want = "Section one"
	contentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	comparePart(t, p, wantp)

	want = "Section two"
	contentContainsString(t, p.Content, want)
}

func TestMultiOtherParts(t *testing.T) {
	var want string
	var wantp *Part
	r := openTestData("parts", "multiother.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		FirstChild:  partExists,
		ContentType: "multipart/x-enmime",
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	comparePart(t, p, wantp)

	want = "Section one"
	contentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	comparePart(t, p, wantp)

	want = "Section two"
	contentContainsString(t, p.Content, want)
}

func TestNestedAlternParts(t *testing.T) {
	var want string
	var wantp *Part
	r := openTestData("parts", "nestedmulti.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		ContentType: "multipart/alternative",
		FirstChild:  partExists,
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	comparePart(t, p, wantp)

	want = "A text section"
	contentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		FirstChild:  partExists,
		ContentType: "multipart/related",
		PartID:      "2.0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// First nested
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2.1",
	}
	comparePart(t, p, wantp)

	want = "An HTML section"
	contentContainsString(t, p.Content, want)

	// Second nested
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Disposition: "inline",
		FileName:    "attach.txt",
		PartID:      "2.2",
	}
	comparePart(t, p, wantp)

	want = "An inline text attachment"
	contentContainsString(t, p.Content, want)

	// Third nested
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/plain",
		Disposition: "inline",
		FileName:    "attach2.txt",
		PartID:      "2.3",
	}
	comparePart(t, p, wantp)

	want = "Another inline text attachment"
	contentContainsString(t, p.Content, want)
}

func TestPartSimilarBoundary(t *testing.T) {
	var want string
	var wantp *Part
	r := openTestData("parts", "similar-boundary.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		ContentType: "multipart/mixed",
		FirstChild:  partExists,
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	comparePart(t, p, wantp)

	want = "Section one"
	contentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		FirstChild:  partExists,
		ContentType: "multipart/alternative",
		PartID:      "2.0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// First nested
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "2.1",
	}
	comparePart(t, p, wantp)

	want = "A text section"
	contentContainsString(t, p.Content, want)

	// Second nested
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2.2",
	}
	comparePart(t, p, wantp)

	want = "An HTML section"
	contentContainsString(t, p.Content, want)
}

// Check we don't UTF-8 decode attachments
func TestBinaryDecode(t *testing.T) {
	var want string
	var wantp *Part
	r := openTestData("parts", "bin-attach.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		FirstChild:  partExists,
		ContentType: "multipart/mixed",
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	comparePart(t, p, wantp)

	want = "A text section"
	contentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "application/octet-stream",
		Charset:     "us-ascii",
		Disposition: "attachment",
		FileName:    "test.bin",
		PartID:      "2",
	}
	comparePart(t, p, wantp)

	wantBytes := []byte{
		0x50, 0x4B, 0x03, 0x04, 0x14, 0x00, 0x08, 0x00,
		0x08, 0x00, 0xC2, 0x02, 0x29, 0x4A, 0x00, 0x00}
	contentEqualsBytes(t, p.Content, wantBytes)
}

func TestMultiBase64Parts(t *testing.T) {
	var want string
	var wantp *Part
	r := openTestData("parts", "multibase64.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		FirstChild:  partExists,
		ContentType: "multipart/mixed",
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	contentEqualsString(t, p.Content, "")

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	comparePart(t, p, wantp)

	want = "A text section"
	contentContainsString(t, p.Content, want)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/html",
		Disposition: "attachment",
		FileName:    "test.html",
		PartID:      "2",
	}
	comparePart(t, p, wantp)

	want = "<html>"
	contentContainsString(t, p.Content, want)
}

func TestBadBoundaryTerm(t *testing.T) {
	var want string
	var wantp *Part
	r := openTestData("parts", "badboundary.raw")
	p, err := ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatal("Unexpected parse error:", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	wantp = &Part{
		FirstChild:  partExists,
		ContentType: "multipart/alternative",
		PartID:      "0",
	}
	comparePart(t, p, wantp)

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	comparePart(t, p, wantp)

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	comparePart(t, p, wantp)

	want = "An HTML section"
	contentContainsString(t, p.Content, want)
}
