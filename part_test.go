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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		Charset:     "ISO-8859-1",
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
		Charset:     "ISO-8859-1",
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		Charset:     "ISO-8859-1",
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
		Charset:     "ISO-8859-1",
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		Charset:     "ISO-8859-1",
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		t.Fatalf("Unexpected parse error: %+v", err)
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
		NextSibling: test.PartExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
		PartID:      "2",
	}
	test.ComparePart(t, p, wantp)

	want = "An HTML section"
	test.ContentContainsString(t, p.Content, want)
}

func TestClonePart(t *testing.T) {
	r := test.OpenTestData("parts", "multiother.raw")
	p, err := enmime.ReadParts(r)

	// Examine root
	if err != nil {
		t.Fatalf("Unexpected parse error: %+v", err)
	}
	if p == nil {
		t.Fatal("Root node should not be nil")
	}

	clone := p.Clone(nil)
	test.ComparePart(t, clone, p)
}

func TestBarrenContentType(t *testing.T) {
	r := test.OpenTestData("parts", "barren-content-type.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	wantp := &enmime.Part{
		PartID:      "0",
		Disposition: "attachment",
	}
	test.ComparePart(t, p, wantp)

	expected := enmime.ErrorMissingContentType
	satisfied := false
	for _, perr := range p.Errors {
		if perr.Name == expected {
			satisfied = true
			if perr.Severe {
				t.Errorf("Expected Severe to be false, got true for %q", perr.Name)
			}
		}
	}
	if !satisfied {
		t.Errorf(
			"Did not find expected error on part. Expected %q, but had: %v",
			expected,
			p.Errors)
	}
}

func TestEmptyContentTypeBadContent(t *testing.T) {
	r := test.OpenTestData("parts", "empty-ctype-bad-content.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	wantp := &enmime.Part{
		PartID:      "1",
		Parent:      test.PartExists,
		Disposition: "",
	}
	test.ComparePart(t, p.FirstChild, wantp)

	expected := enmime.ErrorMissingContentType
	satisfied := false
	for _, perr := range p.FirstChild.Errors {
		if perr.Name == expected {
			satisfied = true
			if perr.Severe {
				t.Errorf("Expected Severe to be false, got true for %q", perr.Name)
			}
		}
	}
	if !satisfied {
		t.Errorf(
			"Did not find expected error on part. Expected %q, but had: %v",
			expected,
			p.Errors)
	}
}

func TestMalformedContentTypeParams(t *testing.T) {
	r := test.OpenTestData("parts", "malformed-content-type-params.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	wantp := &enmime.Part{
		PartID:      "0",
		ContentType: "text/html",
	}
	test.ComparePart(t, p, wantp)

	expected := enmime.ErrorMalformedHeader
	satisfied := false
	for _, perr := range p.Errors {
		if perr.Name == expected {
			satisfied = true
			if perr.Severe {
				t.Errorf("Expected Severe to be false, got true for %q", perr.Name)
			}
		}
	}
	if !satisfied {
		t.Errorf(
			"Did not find expected error on part. Expected %q, but had: %v",
			expected,
			p.Errors)
	}
}

func TestContentTypeParamUnquotedSpecial(t *testing.T) {
	r := test.OpenTestData("parts", "unquoted-ctype-param-special.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	wantp := &enmime.Part{
		PartID:      "0",
		ContentType: "text/calendar",
		Disposition: "attachment",
		FileName:    "calendar.ics",
	}
	test.ComparePart(t, p, wantp)
}

func TestNoClosingBoundary(t *testing.T) {
	r := test.OpenTestData("parts", "multimixed-no-closing-boundary.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if p == nil {
		t.Fatal("Expected part but got nil")
	}

	wantp := &enmime.Part{
		Parent:      test.PartExists,
		PartID:      "1",
		ContentType: "text/html",
		Charset:     "UTF-8",
	}
	test.ComparePart(t, p.FirstChild, wantp)
	t.Log(string(p.FirstChild.Content))

	expected := "Missing Boundary"
	hasCorrectError := false
	for _, v := range p.Errors {
		if v.Severe {
			t.Errorf("Expected Severe to be false, got true for %q", v.Name)
		}
		if v.Name == expected {
			hasCorrectError = true
		}
	}
	if !hasCorrectError {
		t.Fatalf("Did not find expected error on part. Expected %q but got %v", expected, p.Errors)
	}
}

func TestContentTypeParamMissingClosingQuote(t *testing.T) {
	r := test.OpenTestData("parts", "missing-closing-param-quote.raw")
	p, err := enmime.ReadParts(r)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	wantp := &enmime.Part{
		PartID:      "0",
		ContentType: "text/html",
		Charset:     "UTF-8Return-Path: bounce-810_HTML-769869545-477063-1070564-43@bounce.email.oflce57578375.com",
	}
	test.ComparePart(t, p, wantp)

	expected := enmime.ErrorCharsetConversion
	satisfied := false
	for _, perr := range p.Errors {
		if perr.Name == expected {
			satisfied = true
			if perr.Severe {
				t.Errorf("Expected Severe to be false, got true for %q", perr.Name)
			}
		}
	}
	if !satisfied {
		t.Errorf(
			"Did not find expected error on part. Expected %q, but had: %v",
			expected,
			p.Errors)
	}
}

func TestChardetFailure(t *testing.T) {
	const expectedContent = "GIF89ad\x00\x04\x00\x80\x00\x00\x00f\xccf\xff\x99!\xf9\x04\x00\x00\x00\x00\x00,\x00\x00\x00\x00d\x00\x04\x00\x00\x02\x1a\x8c\x8f\xa9\xcb\xed\x0f\xa3\x9c\xb4\xda\xeb\x80\u07bc\xfb\x0f\x86\xe2H\x96扦\xea*\x16\x00;"

	t.Run("text part", func(t *testing.T) {
		r := test.OpenTestData("parts", "chardet-fail.raw")
		p, err := enmime.ReadParts(r)
		if err != nil {
			t.Fatal(err)
		}
		wantp := &enmime.Part{
			PartID:      "0",
			ContentType: "text/plain",
			ContentID:   "part3.E34FF3C4.059DAD00@example.com",
			FileName:    "rzkly.txt",
		}
		test.ComparePart(t, p, wantp)
		expected := enmime.ErrorCharsetDeclaration
		satisfied := false
		for _, perr := range p.Errors {
			if perr.Name == expected {
				satisfied = true
				if perr.Severe {
					t.Errorf("Expected Severe to be false, got true for %q", perr.Name)
				}
			}
		}
		if !satisfied {
			t.Errorf(
				"Did not find expected error on part. Expected %q, but had: %v",
				expected,
				p.Errors)
		}
		test.ContentEqualsString(t, p.Content, expectedContent)
	})

	t.Run("non-text part", func(t *testing.T) {
		r := test.OpenTestData("parts", "chardet-fail-non-txt.raw")
		p, err := enmime.ReadParts(r)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Errors) > 0 {
			t.Errorf("Errors encountered while processing part: %v", p.Errors)
		}
		wantp := &enmime.Part{
			PartID:      "0",
			ContentType: "image/gif",
			ContentID:   "part3.E34FF3C4.059DAD00@example.com",
			FileName:    "rzkly.gif",
		}
		test.ComparePart(t, p, wantp)
		test.ContentEqualsString(t, p.Content, expectedContent)
	})
}
