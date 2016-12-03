package enmime

import (
	"io/ioutil"
	"strings"
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
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if got != want {
		t.Errorf("ContentType got: %q, want: %q", got, want)
	}
}

func TestMultiAlternParts(t *testing.T) {
	var want, got string
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

	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content should have length of 0")
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
	}
}

func TestPartMissingContentType(t *testing.T) {
	var want, got string
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

	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content should have length of 0")
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
	}
}

func TestPartEmptyHeader(t *testing.T) {
	var want, got string
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

	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content should have length of 0")
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
	}
}

func TestMultiMixedParts(t *testing.T) {
	var want, got string
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

	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content should have length of 0")
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
	}
}

func TestMultiOtherParts(t *testing.T) {
	var want, got string
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

	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content should have length of 0")
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
	}
}

func TestNestedAlternParts(t *testing.T) {
	var want, got string
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
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content should have length of 0")
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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

	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content should have length of 0")
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
	}
}

func TestMultiBase64Parts(t *testing.T) {
	var want, got string
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

	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(allBytes) > 0 {
		t.Error("Content should have length of 0")
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
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
	allBytes, err = ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
	}
}

func TestBadBoundaryTerm(t *testing.T) {
	var want, got string
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
	allBytes, err := ioutil.ReadAll(p)
	if err != nil {
		t.Fatal(err)
	}
	got = string(allBytes)
	if !strings.Contains(got, want) {
		t.Errorf("Content: %q, should contain: %q", got, want)
	}
}
