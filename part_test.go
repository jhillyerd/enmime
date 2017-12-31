package enmime

import (
	"bufio"
	"bytes"
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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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
	if ok, err := contentEqualsString(p, want); !ok {
		t.Error("Part", err)
	}
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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		FirstChild:  partExists,
		ContentType: "multipart/related",
		PartID:      "2.0",
	}
	comparePart(t, p, wantp)

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		FirstChild:  partExists,
		ContentType: "multipart/alternative",
		PartID:      "2.0",
	}
	comparePart(t, p, wantp)

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentEqualsBytes(p, wantBytes); !ok {
		t.Error("Part", err)
	}
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

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}

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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
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
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
}

func TestSplitEpiloge(t *testing.T) {
	emailBody := bytes.NewBuffer([]byte(`--Enmime-Test-100
Content-Transfer-Encoding: 7bit
Content-Type: text/plain; charset=us-ascii

A text section
--Enmime-Test-100
Content-Transfer-Encoding: base64
Content-Type: text/html; name="test.html"
Content-Disposition: attachment; filename=test.html

PGh0bWw+Cg==

--Enmime-Test-100-->
PGh0bWw+Cg==`))
	boundary := "Enmime-Test-100"
	expectedBody := bytes.NewBuffer([]byte(`--Enmime-Test-100
Content-Transfer-Encoding: 7bit
Content-Type: text/plain; charset=us-ascii

A text section
--Enmime-Test-100
Content-Transfer-Encoding: base64
Content-Type: text/html; name="test.html"
Content-Disposition: attachment; filename=test.html

PGh0bWw+Cg==

--Enmime-Test-100-->
`))
	expectedEpilogue := bytes.NewBuffer([]byte(`PGh0bWw+Cg==`))

	body, epilogue, err := splitEpilogue(bufio.NewReader(emailBody), boundary)
	if err != nil {
		t.Error("Error getting epilogue", err)
	}
	if !bytes.Equal(body.Bytes(), expectedBody.Bytes()) {
		t.Error("Error mismatch body")
	}
	if !bytes.Equal(epilogue.Bytes(), expectedEpilogue.Bytes()) {
		t.Error("Error mismatch epilogue")
	}
}
