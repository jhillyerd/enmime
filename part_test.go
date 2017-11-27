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
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

	want = "An HTML section"
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		// No ContentType
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

	if ok, err := contentEqualsString(p, ""); !ok {
		t.Error("Part", err)
	}

	// Examine first child
	p = p.FirstChild

	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		// No ContentType
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

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
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

	// Examine first child
	p = p.FirstChild
	wantp = &Part{
		Parent:      partExists,
		NextSibling: partExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

	// Examine sibling
	p = p.NextSibling
	wantp = &Part{
		Parent:      partExists,
		ContentType: "text/html",
		Charset:     "us-ascii",
	}
	comparePart(p, wantp, func(field, got, want string) {
		t.Errorf("Part.%s == %q, want: %q", field, got, want)
	})

	want = "An HTML section"
	if ok, err := contentContainsString(p, want); !ok {
		t.Error("Part", err)
	}
}
