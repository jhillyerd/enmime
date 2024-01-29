package enmime_test

import (
	"strings"
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

func TestAddChildInfiniteLoops(t *testing.T) {
	// Part adds itself
	parentPart := &enmime.Part{
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "0",
	}
	parentPart.AddChild(parentPart)

	// Part adds its own FirstChild
	childPart := &enmime.Part{
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "1",
	}
	parentPart.FirstChild = childPart
	parentPart.AddChild(childPart)

	parentPart.FirstChild = nil
	// Part adds a child that is its own NextSibling
	childPart.NextSibling = childPart
	parentPart.AddChild(childPart)
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

func TestMultiSkipMalformedPart(t *testing.T) {
	var want string
	var wantp *enmime.Part
	r := test.OpenTestData("parts", "multi-malformed.raw")
	parser := enmime.NewParser(enmime.SkipMalformedParts(true))
	p, err := parser.ReadParts(r)

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

	want = "Section one"
	test.ContentContainsString(t, p.Content, want)

	// Verify sibling is Section two and not the malformed part.
	p = p.NextSibling
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		ContentType: "text/plain",
		Charset:     "us-ascii",
		PartID:      "3",
	}
	test.ComparePart(t, p, wantp)

	want = "Section two"
	test.ContentContainsString(t, p.Content, want)
}

func TestReadPartErrorPolicy(t *testing.T) {
	// example policy 1
	examplePolicy1 := enmime.AllowCorruptTextPartErrorPolicy

	// example policy 2: recover partial content from base64.CorruptInputError when content type is text/plain
	examplePolicy2 := enmime.ReadPartErrorPolicy(func(p *enmime.Part, err error) bool {
		if enmime.IsBase64CorruptInputError(err) && p.ContentType == "text/plain" {
			return true
		}
		return false
	})

	// example policy 3: always recover the partial content read, no matter the error
	examplePolicy3 := enmime.ReadPartErrorPolicy(func(p *enmime.Part, err error) bool {
		return true
	})

	plainContent := "Hello World!"
	htmlContent := "<p>Hello World</p>"

	illegalBase64Error := enmime.Error{
		Name:   enmime.ErrorMalformedBase64,
		Detail: "illegal base64 data at input byte 0",
		Severe: true,
	}

	keepingBufferWarning := enmime.Error{
		Name:   enmime.ErrorMalformedChildPart,
		Detail: "partial content: illegal base64 data at input byte 0",
		Severe: false,
	}

	tests := map[string]struct {
		policy                       enmime.ReadPartErrorPolicy
		expectedPlainTextPartContent string
		expectedHTMLPartContent      string
		expectedErrorPlainTextPart   enmime.Error
		expectedErrorHTMLPart        enmime.Error
	}{
		"no policy (default behavior)": {
			policy:                       nil,
			expectedPlainTextPartContent: "",
			expectedErrorPlainTextPart:   illegalBase64Error,
			expectedHTMLPartContent:      "",
			expectedErrorHTMLPart:        illegalBase64Error,
		},
		"keep buffer according to policy 1": {
			policy:                       examplePolicy1,
			expectedPlainTextPartContent: plainContent,
			expectedErrorPlainTextPart:   keepingBufferWarning,
			expectedHTMLPartContent:      htmlContent,
			expectedErrorHTMLPart:        keepingBufferWarning,
		},
		"keep buffer according to policy 2": {
			policy:                       examplePolicy2,
			expectedPlainTextPartContent: plainContent,
			expectedErrorPlainTextPart:   keepingBufferWarning,
			expectedHTMLPartContent:      "",
			expectedErrorHTMLPart:        illegalBase64Error,
		},
		"keep buffer according to policy 3": {
			policy:                       examplePolicy3,
			expectedPlainTextPartContent: plainContent,
			expectedErrorPlainTextPart:   keepingBufferWarning,
			expectedHTMLPartContent:      htmlContent,
			expectedErrorHTMLPart:        keepingBufferWarning,
		},
	}

	for testName, testData := range tests {
		t.Run(testName, func(t *testing.T) {
			r := test.OpenTestData("parts", "extra-base64-character.raw")
			parser := enmime.NewParser(enmime.SetReadPartErrorPolicy(testData.policy))
			p, err := parser.ReadParts(r)

			if err != nil {
				t.Fatalf("Unexpected parse error: %+v", err)
			}

			var expectedPart *enmime.Part

			// Examine root
			expectedPart = &enmime.Part{
				FirstChild:  test.PartExists,
				ContentType: "multipart/alternative",
				PartID:      "0",
			}
			expectedSubject := "Each part contains 1 extra base64 character (4*n + 1)"
			gotSubject := p.Header.Get("Subject")

			test.ComparePart(t, p, expectedPart)
			test.ContentEqualsString(t, p.Content, "")
			if gotSubject != expectedSubject {
				t.Errorf("Subject got: %q, expected: %q", gotSubject, expectedSubject)
			}

			// Examine parts
			expectedCharset := "utf-8"
			expectedContentTransferEncoding := "base64"
			var foundExpectedErr bool

			// Examine first child
			p = p.FirstChild
			expectedPart = &enmime.Part{
				Parent:      test.PartExists,
				NextSibling: test.PartExists,
				ContentType: "text/plain",
				Charset:     expectedCharset,
				PartID:      "1",
			}
			gotContentTransferEncoding := p.Header.Get("Content-Transfer-Encoding")

			test.ComparePart(t, p, expectedPart)
			test.ContentEqualsString(t, p.Content, testData.expectedPlainTextPartContent)
			if gotContentTransferEncoding != expectedContentTransferEncoding {
				t.Errorf("Content-Transfer-Encoding got: %q, expected: %q", gotContentTransferEncoding, expectedContentTransferEncoding)
			}
			foundExpectedErr = false
			for _, v := range p.Errors {
				if *v == testData.expectedErrorPlainTextPart {
					foundExpectedErr = true
					break
				}
			}
			if !foundExpectedErr {
				t.Errorf("Expected to find error: %v", testData.expectedErrorPlainTextPart)
			}

			// Examine second child
			p = p.NextSibling
			expectedPart = &enmime.Part{
				Parent:      test.PartExists,
				ContentType: "text/html",
				Charset:     expectedCharset,
				PartID:      "2",
			}
			gotContentTransferEncoding = p.Header.Get("Content-Transfer-Encoding")

			test.ComparePart(t, p, expectedPart)
			test.ContentEqualsString(t, p.Content, testData.expectedHTMLPartContent)
			if gotContentTransferEncoding != expectedContentTransferEncoding {
				t.Errorf("Content-Transfer-Encoding got: %q, expected: %q", gotContentTransferEncoding, expectedContentTransferEncoding)
			}
			foundExpectedErr = false
			for _, v := range p.Errors {
				if *v == testData.expectedErrorHTMLPart {
					foundExpectedErr = true
					break
				}
			}
			if !foundExpectedErr {
				t.Errorf("Expected to find error: %v", testData.expectedErrorHTMLPart)
			}
		})
	}
}

func TestMultiNoSkipMalformedPartFails(t *testing.T) {
	r := test.OpenTestData("parts", "multi-malformed.raw")
	parser := enmime.NewParser(enmime.SkipMalformedParts(false))
	_, err := parser.ReadParts(r)
	if err == nil {
		t.Fatal("Expecting parsing to fail")
	}

	if !strings.Contains(err.Error(), "malformed MIME header initial line") {
		t.Fatal("Expecting for error to contain \"malformed MIME header\" error")
	}
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

	t.Run("not enough characters part", func(t *testing.T) {
		r := test.OpenTestData("parts", "chardet-fail-not-long-enough.raw")
		p, err := enmime.ReadParts(r)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Errors) > 0 {
			t.Errorf("Errors encountered while processing part: %v", p.Errors)
		}
		wantp := &enmime.Part{
			PartID:      "0",
			ContentType: "text/plain",
			Charset:     "UTF-8",
		}
		test.ComparePart(t, p, wantp)
		test.ContentEqualsString(t, p.Content, "和弟弟\r\n")
	})
}

func TestChardetSuccess(t *testing.T) {
	// Testdata in these tests licensed under CC0: Public Domain
	t.Run("big-5 data in us-ascii part", func(t *testing.T) {
		r := test.OpenTestData("parts", "chardet-success-big-5.raw")
		p, err := enmime.ReadParts(r)
		if err != nil {
			t.Fatal(err)
		}
		expectedErr := enmime.Error{
			Name:   "Character Set Declaration Mismatch",
			Detail: "declared charset \"us-ascii\", detected \"Big5\", confidence 100",
			Severe: false,
		}
		foundExpectedErr := false
		if len(p.Errors) > 0 {
			for _, v := range p.Errors {
				if *v == expectedErr {
					foundExpectedErr = true
				} else {
					t.Errorf("Error encountered while processing part: %v", v)
				}
			}
		}
		if !foundExpectedErr {
			t.Errorf("Expected to find %v warning", expectedErr)
		}
		wantp := &enmime.Part{
			PartID:      "0",
			ContentType: "text/plain",
			Charset:     "Big5",
		}
		test.ComparePart(t, p, wantp)
	})

	t.Run("iso-8859-1 data in us-ascii part", func(t *testing.T) {
		r := test.OpenTestData("parts", "chardet-success-iso-8859-1.raw")
		p, err := enmime.ReadParts(r)
		if err != nil {
			t.Fatal(err)
		}
		expectedErr := enmime.Error{
			Name:   "Character Set Declaration Mismatch",
			Detail: "declared charset \"us-ascii\", detected \"ISO-8859-1\", confidence 90",
			Severe: false,
		}
		foundExpectedErr := false
		if len(p.Errors) > 0 {
			for _, v := range p.Errors {
				if *v == expectedErr {
					foundExpectedErr = true
				} else {
					t.Errorf("Error encountered while processing part: %v", v)
				}
			}
		}
		if !foundExpectedErr {
			t.Errorf("Expected to find %v warning", expectedErr)
		}
		wantp := &enmime.Part{
			PartID:      "0",
			ContentType: "text/plain",
			Charset:     "ISO-8859-1",
		}
		test.ComparePart(t, p, wantp)
	})
}

func TestCtypeInvalidCharacters(t *testing.T) {
	r := test.OpenTestData("parts", "ctype-invalid-characters.raw")
	parser := enmime.NewParser(enmime.StripMediaTypeInvalidCharacters(true))
	p, err := parser.ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	wantp := &enmime.Part{
		PartID:      "0",
		ContentType: "text/plain",
	}

	test.ComparePart(t, p, wantp)
}
