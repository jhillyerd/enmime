package enmime_test

import (
	"bytes"
	"fmt"
	"net/mail"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/jhillyerd/enmime/v2"
	"github.com/jhillyerd/enmime/v2/internal/test"
)

func TestParseHeaderOnly(t *testing.T) {
	want := ""

	msg := test.OpenTestData("mail", "header-only.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}
	if !strings.Contains(e.Text, want) {
		t.Errorf("Expected %q to contain %q", e.Text, want)
	}
	if e.HTML != "" {
		t.Errorf("Expected no HTML body, got %q", e.HTML)
	}
	if e.Root == nil {
		t.Error("Expected a root part")
	}
	if len(e.Root.Header) != 7 {
		t.Errorf("Expected 7 headers, got %d", len(e.Root.Header))
	}
	dt, err := e.Date()
	if err != nil {
		t.Errorf("Failed to parse Date header: %v", err)
	} else if !dt.Equal(time.Date(2012, time.October, 19, 5, 48, 39, 0, time.UTC)) {
		t.Errorf("Incorrect date parsed: %s", dt)
	}
}

func TestParseNonMime(t *testing.T) {
	want := "This is a test mailing"
	msg := test.OpenTestData("mail", "non-mime.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}
	if !strings.Contains(e.Text, want) {
		t.Errorf("Expected %q to contain %q", e.Text, want)
	}
	if e.HTML != "" {
		t.Errorf("Expected no HTML body, got %q", e.HTML)
	}
}

func TestParseNonMimeHTML(t *testing.T) {
	msg := test.OpenTestData("mail", "non-mime-html.raw")

	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}
	if len(e.Errors) == 1 {
		want := enmime.ErrorPlainTextFromHTML
		got := e.Errors[0].Name
		if got != want {
			t.Errorf("e.Errors[0] got: %v, want: %v", got, want)
		}
	} else {
		t.Errorf("len(e.Errors) got: %v, want: 1", len(e.Errors))
	}

	want := "This is *a* *test* mailing"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Expected %q to contain %q", e.Text, want)
	}

	want = "<span>This</span>"
	if !strings.Contains(e.HTML, want) {
		t.Errorf("Expected %q to contain %q", e.HTML, want)
	}
}

// TestEnvelopeFromPartDoesNotGatherErrors verifies that EnvelopeFromPart does not collect errors,
// while GatherNestedErrors does. This enforces the separation of concerns: EnvelopeFromPart builds
// the envelope structure, and GatherNestedErrors is responsible for error collection.
func TestEnvelopeFromPartDoesNotGatherErrors(t *testing.T) {
	// Use a message that generates errors (HTML without plain text)
	msg := test.OpenTestData("mail", "non-mime-html.raw")
	parser := enmime.NewParser()
	root, err := parser.ReadParts(msg)
	if err != nil {
		t.Fatalf("Failed to read parts: %v", err)
	}

	// Call EnvelopeFromPart directly (not ReadEnvelope)
	e, err := parser.EnvelopeFromPart(root)
	if err != nil {
		t.Fatalf("Failed to create envelope from part: %v", err)
	}

	// EnvelopeFromPart should NOT have collected errors
	if len(e.Errors) != 0 {
		t.Errorf("EnvelopeFromPart should not gather errors, but got %d errors:", len(e.Errors))
		for _, err := range e.Errors {
			t.Logf("  - %v", err)
		}
	}

	// Verify that the envelope structure was built correctly
	if e.HTML == "" {
		t.Error("Expected HTML content from envelope structure building")
	}

	// Now call GatherNestedErrors
	err = e.GatherNestedErrors()
	if err != nil {
		t.Fatalf("GatherNestedErrors failed: %v", err)
	}

	// GatherNestedErrors SHOULD have collected the error
	if len(e.Errors) != 1 {
		t.Errorf("GatherNestedErrors should collect exactly 1 error, got %d", len(e.Errors))
	} else if e.Errors[0].Name != enmime.ErrorPlainTextFromHTML {
		t.Errorf("Expected error %q, got %q", enmime.ErrorPlainTextFromHTML, e.Errors[0].Name)
	}

	// Verify text was downconverted from HTML
	want := "This is *a* *test* mailing"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Expected %q to contain %q", e.Text, want)
	}
}

func TestParseMimeTree(t *testing.T) {
	msg := test.OpenTestData("mail", "attachment.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if e.Root == nil {
		t.Error("Message should have a root node")
	}
}

func TestParseInlineText(t *testing.T) {
	msg := test.OpenTestData("mail", "html-mime-inline.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Test of text section"
	if e.Text != want {
		t.Error("got:", e.Text, "want:", want)
	}
}

func TestParseInlineBadCharsetText(t *testing.T) {
	msg := test.OpenTestData("mail", "html-mime-bad-charset-inline.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Test of text section"
	if e.Text != want {
		t.Error("got:", e.Text, "want:", want)
	}
}

func TestParseInlineBadUknownCharsetText(t *testing.T) {
	msg := test.OpenTestData("mail", "html-mime-bad-unknown-charset-inline.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Test of text section"
	if e.Text != want {
		t.Error("got:", e.Text, "want:", want)
	}
}

func TestParseMultiAlernativeText(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-alternative.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Section one\n"
	if e.Text != want {
		t.Error("Text parts should not concatenate, got:", e.Text, "want:", want)
	}
}

func TestParseMultiMixedText(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-mixed.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Section one\n\n--\nSection two"
	if e.Text != want {
		t.Error("Text parts should concatenate, got:", e.Text, "want:", want)
	}
}

// TestParseDelimiterMidLine checks that the boundary is only taken as a
// delimiter at the start of a line (RFC 2046 5.1.1): a part whose body or
// headers mention "--boundary" elsewhere on a line is still one part, and a
// quoted-printable body that encodes to such a line is decoded whole.
func TestParseDelimiterMidLine(t *testing.T) {
	msg := "Content-Type: multipart/mixed; boundary=bnd\r\n\r\n" +
		"--bnd\r\nContent-Type: text/html\r\n\r\n<p>hi</p>\r\n --bnd\r\nContent-Type: text/plain\r\n\r\nnot a part\r\n" +
		"--bnd\r\nContent-Type: text/plain\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\nhello\r\n=20--bnd\r\nmore\r\n" +
		"--bnd--\r\n"
	e, err := enmime.ReadEnvelope(strings.NewReader(msg))
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	var kids []*enmime.Part
	for p := e.Root.FirstChild; p != nil; p = p.NextSibling {
		kids = append(kids, p)
	}
	if len(kids) != 2 {
		t.Fatalf("got %d parts, want 2", len(kids))
	}
	if got, want := string(kids[0].Content), "<p>hi</p>\r\n --bnd\r\nContent-Type: text/plain\r\n\r\nnot a part"; got != want {
		t.Errorf("part 1 content got %q, want %q", got, want)
	}
	if got, want := string(kids[1].Content), "hello\r\n --bnd\r\nmore"; got != want {
		t.Errorf("part 2 content got %q, want %q", got, want)
	}
}

func TestParseMultiMixedRelatedHtml(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-mixed-related.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "<p>html1</p>"
	if e.HTML != want {
		t.Error("Text parts should concatenate, got:", e.HTML, "want:", want)
	}
}

func TestParseMultiSignedText(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-signed.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Section one\n\n--\nSection two"
	if e.Text != want {
		t.Error("Text parts should concatenate, got:", e.Text, "want:", want)
	}
}

func TestParseQuotedPrintable(t *testing.T) {
	msg := test.OpenTestData("mail", "quoted-printable.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Phasellus sit amet arcu"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Text: %q should contain: %q", e.Text, want)
	}
}

func TestParseQuotedPrintableMime(t *testing.T) {
	msg := test.OpenTestData("mail", "quoted-printable-mime.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Nullam venenatis ante"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Text: %q should contain: %q", e.Text, want)
	}
}

func TestParseInlineHTML(t *testing.T) {
	msg := test.OpenTestData("mail", "html-mime-inline.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "<html>"
	if !strings.Contains(e.HTML, want) {
		t.Errorf("HTML: %q should contain: %q", e.Text, want)
	}

	want = "Test of HTML section"
	if !strings.Contains(e.HTML, want) {
		t.Errorf("HTML: %q should contain: %q", e.Text, want)
	}
}

func TestParseAttachment(t *testing.T) {
	msg := test.OpenTestData("mail", "attachment.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "A text section"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Text: %q should contain: %q", e.Text, want)
	}
	if e.HTML != "" {
		t.Error("mime.HTML should be empty, attachment is not for display, got:", e.HTML)
	}
	if len(e.Inlines) > 0 {
		t.Error("Should have no inlines, got:", len(e.Inlines))
	}
	if len(e.Attachments) != 1 {
		t.Fatal("Should have a single attachment, got:", len(e.Attachments))
	}

	want = "test.html"
	got := e.Attachments[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}

	want = "<html>"
	test.ContentContainsString(t, e.Attachments[0].Content, want)
}

func TestParseAttachmentOctet(t *testing.T) {
	msg := test.OpenTestData("mail", "attachment-octet.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "A text section"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Text: %q should contain: %q", e.Text, want)
	}
	if e.HTML != "" {
		t.Error("mime.HTML should be empty, attachment is not for display, got:", e.HTML)
	}
	if len(e.Inlines) > 0 {
		t.Error("Should have no inlines, got:", len(e.Inlines))
	}
	if len(e.Attachments) != 1 {
		t.Fatal("Should have a single attachment, got:", len(e.Attachments))
	}

	want = "ATTACHMENT.EXE"
	got := e.Attachments[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}

	wantBytes := []byte{
		0x3, 0x17, 0xe1, 0x7e, 0xe8, 0xeb, 0xa2, 0x96, 0x9d, 0x95, 0xa7, 0x67, 0x82, 0x9,
		0xdf, 0x8e, 0xc, 0x2c, 0x6a, 0x2b, 0x9b, 0xbe, 0x79, 0xa4, 0x69, 0xd8, 0xae, 0x86,
		0xd7, 0xab, 0xa8, 0x72, 0x52, 0x15, 0xfb, 0x80, 0x8e, 0x47, 0xe1, 0xae, 0xaa, 0x5e,
		0xa2, 0xb2, 0xc0, 0x90, 0x59, 0xe3, 0x35, 0xf8, 0x60, 0xb7, 0xb1, 0x63, 0x77, 0xd7,
		0x5f, 0x92, 0x58, 0xa8, 0x75,
	}
	if !bytes.Equal(e.Attachments[0].Content, wantBytes) {
		t.Error("Attachment should have correct content")
	}
}

func TestParseAttachmentApplication(t *testing.T) {
	msg := test.OpenTestData("mail", "attachment-application.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if len(e.Inlines) > 0 {
		t.Error("Should have no inlines, got:", len(e.Inlines))
	}
	if len(e.Attachments) != 1 {
		t.Fatal("Should have a single attachment, got:", len(e.Attachments))
	}

	want := "some.doc"
	got := e.Attachments[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}
}

// TestParseAttachmentMissingContentType expects Content-Disposition and Content-ID to be read
// from parts that have no Content-Type header.
func TestParseAttachmentMissingContentType(t *testing.T) {
	msg := test.OpenTestData("mail", "attachment-missing-ctype.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "A text section"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Text: %q should contain: %q", e.Text, want)
	}

	if len(e.Attachments) != 1 {
		t.Fatal("Should have a single attachment, got:", len(e.Attachments))
	}
	wantp := &enmime.Part{
		Parent:      test.PartExists,
		NextSibling: test.PartExists,
		// No ContentType
		Disposition: "attachment",
		FileName:    "notes.txt",
		PartID:      "2",
	}
	test.ComparePart(t, e.Attachments[0], wantp)

	if len(e.Inlines) != 1 {
		t.Fatal("Should have a single inline, got:", len(e.Inlines))
	}
	wantp = &enmime.Part{
		Parent:      test.PartExists,
		Disposition: "inline",
		PartID:      "3",
	}
	test.ComparePart(t, e.Inlines[0], wantp)
	want = "part123456@inbucket.org"
	if got := e.Inlines[0].ContentID; got != want {
		t.Errorf("ContentID got: %q, want: %q", got, want)
	}

	if len(e.OtherParts) > 0 {
		t.Error("Should have no other parts, got:", len(e.OtherParts))
	}

	// Two warnings for each of the two parts, as before.
	if len(e.Errors) != 4 {
		t.Errorf("len(e.Errors) got: %v, want: 4", len(e.Errors))
	}
	for _, perr := range e.Errors {
		if perr.Name != enmime.ErrorMissingContentType {
			t.Errorf("e.Errors got: %v, want only: %v", perr.Name, enmime.ErrorMissingContentType)
		}
	}
}

func TestParseOtherParts(t *testing.T) {
	msg := test.OpenTestData("mail", "other-parts.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "A text section"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Text: %q should contain: %q", e.Text, want)
	}
	if e.HTML != "" {
		t.Error("mime.HTML should be empty, attachment is not for display, got:", e.HTML)
	}
	if len(e.Inlines) > 0 {
		t.Error("Should have no inlines, got:", len(e.Inlines))
	}
	if len(e.Attachments) > 0 {
		t.Error("Should have no attachments, got:", len(e.Attachments))
	}
	if len(e.OtherParts) != 1 {
		t.Fatal("Should have one other part, got:", len(e.OtherParts))
	}

	want = "B05.gif"
	got := e.OtherParts[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}
	wantBytes := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0xf, 0x0, 0xf, 0x0, 0xa2, 0x5, 0x0, 0xde, 0xeb,
		0xf3, 0x5b, 0xb0, 0xec, 0x0, 0x89, 0xe3, 0xa3, 0xd0, 0xed, 0x0, 0x46, 0x74, 0xdd,
		0xed, 0xfa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x21, 0xf9, 0x4, 0x1, 0x0, 0x0, 0x5, 0x0,
		0x2c, 0x0, 0x0, 0x0, 0x0, 0xf, 0x0, 0xf, 0x0, 0x0, 0x3, 0x40, 0x58, 0x25, 0xa4, 0x4b,
		0xb0, 0x39, 0x1, 0x46, 0xa3, 0x23, 0x5b, 0x47, 0x46, 0x68, 0x9d, 0x20, 0x6, 0x9f,
		0xd2, 0x95, 0x45, 0x44, 0x8, 0xe8, 0x29, 0x39, 0x69, 0xeb, 0xbd, 0xc, 0x41, 0x4a,
		0xae, 0x82, 0xcd, 0x1c, 0x9f, 0xce, 0xaf, 0x1f, 0xc3, 0x34, 0x18, 0xc2, 0x42, 0xb8,
		0x80, 0xf1, 0x18, 0x84, 0xc0, 0x9e, 0xd0, 0xe8, 0xf2, 0x1, 0xb5, 0x19, 0xad, 0x41,
		0x53, 0x33, 0x9b, 0x0, 0x0, 0x3b,
	}
	if !bytes.Equal(e.OtherParts[0].Content, wantBytes) {
		t.Error("Other part should have correct content")
	}
}

func TestParseInline(t *testing.T) {
	msg := test.OpenTestData("mail", "html-mime-inline.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Test of text section"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Text: %q should contain: %q", e.Text, want)
	}

	want = ">Test of HTML section<"
	if !strings.Contains(e.HTML, want) {
		t.Errorf("HTML: %q should contain %q", e.HTML, want)
	}

	if len(e.Inlines) != 1 {
		t.Fatal("Should have one inline, got:", len(e.Inlines))
	}
	if len(e.Attachments) > 0 {
		t.Error("Should have no attachments, got:", len(e.Attachments))
	}

	want = "favicon.png"
	got := e.Inlines[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}
	if !bytes.HasPrefix(e.Inlines[0].Content, []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("Inline should have correct content")
	}
}

func TestParseOtherPartsRelated(t *testing.T) {
	msg := test.OpenTestData("mail", "other-multi-related.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Plain text."
	if !strings.Contains(e.Text, want) {
		t.Errorf("Text: %q should contain: %q", e.Text, want)
	}

	want = "<i>HTML text.</i>"
	if !strings.Contains(e.HTML, want) {
		t.Errorf("HTML: %q should contain %q", e.HTML, want)
	}

	if len(e.Attachments) > 0 {
		t.Error("Should have no attachments, got:", len(e.Attachments))
	}
	if len(e.Inlines) > 0 {
		t.Error("Should have no inlines, got:", len(e.Inlines))
	}
	if len(e.OtherParts) != 2 {
		t.Fatal("Should have two other parts, got:", len(e.Inlines))
	}

	want = "image001.png"
	got := e.OtherParts[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}
	want = "image001.png@01D3BA12.F6C6AEB0"
	got = e.OtherParts[0].ContentID
	if got != want {
		t.Error("ContentID got:", got, "want:", want)
	}
	if !bytes.HasPrefix(e.OtherParts[0].Content, []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("Other part should have correct content")
	}

	want = "image002.png"
	got = e.OtherParts[1].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}
	want = "image002.png@01D3BA12.F6C6AEB0"
	got = e.OtherParts[1].ContentID
	if got != want {
		t.Error("ContentID got:", got, "want:", want)
	}
	if !bytes.HasPrefix(e.OtherParts[1].Content, []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("Other part should have correct content")
	}
}

func TestParseHTMLOnlyInline(t *testing.T) {
	msg := test.OpenTestData("mail", "html-only-inline.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if len(e.Errors) == 1 {
		want := enmime.ErrorPlainTextFromHTML
		got := e.Errors[0].Name
		if got != want {
			t.Errorf("e.Errors[0] got: %v, want: %v", got, want)
		}
	} else {
		t.Errorf("len(e.Errors) got: %v, want: 1", len(e.Errors))
	}

	want := "Test of HTML section"
	if !strings.Contains(e.Text, want) {
		t.Errorf("Downconverted Text: %q should contain: %q", e.Text, want)
	}

	want = ">Test of HTML section<"
	if !strings.Contains(e.HTML, want) {
		t.Errorf("HTML: %q should contain %q", e.HTML, want)
	}

	if len(e.Inlines) != 1 {
		t.Error("Should one inline, got:", len(e.Inlines))
	}
	if len(e.Attachments) > 0 {
		t.Fatal("Should have no attachments, got:", len(e.Attachments))
	}

	want = "favicon.png"
	got := e.Inlines[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}
	if !bytes.HasPrefix(e.Inlines[0].Content, []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("Inline should have correct content")
	}
}

func TestParseHTMLOnlyInlineTextDisabled(t *testing.T) {
	msg := test.OpenTestData("mail", "html-only-inline.raw")

	parser := enmime.NewParser(enmime.DisableTextConversion(true))
	e, err := parser.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if len(e.Errors) == 1 {
		want := enmime.ErrorPlainTextFromHTML
		got := e.Errors[0].Name
		if got != want {
			t.Errorf("e.Errors[0] got: %v, want: %v", got, want)
		}
	} else {
		t.Errorf("len(e.Errors) got: %v, want: 1", len(e.Errors))
	}

	if e.Text != "" {
		t.Errorf("Downconverted Text was %q, should be empty", e.Text)
	}

	want := ">Test of HTML section<"
	if !strings.Contains(e.HTML, want) {
		t.Errorf("HTML: %q should contain %q", e.HTML, want)
	}

	if len(e.Inlines) != 1 {
		t.Error("Should one inline, got:", len(e.Inlines))
	}
	if len(e.Attachments) > 0 {
		t.Fatal("Should have no attachments, got:", len(e.Attachments))
	}

	want = "favicon.png"
	got := e.Inlines[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}
	if !bytes.HasPrefix(e.Inlines[0].Content, []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("Inline should have correct content")
	}
}

func TestParseInlineMultipart(t *testing.T) {
	msg := test.OpenTestData("mail", "inlinemultipart.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if len(e.Errors) != 0 {
		t.Errorf("len(e.Errors) got: %v, want: 0", len(e.Errors))
	}

	want := "Simple text."
	if !strings.Contains(e.Text, want) {
		t.Errorf("Downconverted Text: %q should contain: %q", e.Text, want)
	}

	if len(e.Inlines) != 1 {
		t.Error("Should have one inline, got:", len(e.Inlines))
	}
	if len(e.Attachments) != 1 {
		t.Fatal("Should have one attachments, got:", len(e.Attachments))
	}

	want = "test.txt"
	got := e.Inlines[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}

	if !bytes.HasPrefix(e.Inlines[0].Content, []byte("Text")) {
		t.Error("Inline should have correct content")
	}
}

func TestParseNestedHeaders(t *testing.T) {
	msg := test.OpenTestData("mail", "html-mime-inline.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if len(e.Inlines) != 1 {
		t.Error("Should one inline, got:", len(e.Inlines))
	}

	want := "favicon.png"
	got := e.Inlines[0].FileName
	if got != want {
		t.Error("FileName got:", got, "want:", want)
	}
	want = "<8B8481A2-25CA-4886-9B5A-8EB9115DD064@skynet>"
	got = e.Inlines[0].Header.Get("Content-Id")
	if got != want {
		t.Errorf("Content-Id header was: %q, want: %q", got, want)
	}
}

func TestParseHTMLOnlyCharsetInHeaderOnly(t *testing.T) {
	msg := test.OpenTestData("mail", "non-mime-html-charset-header-only.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}

	if !strings.ContainsRune(e.HTML, 0xfc) {
		t.Error("HTML body should contained German ü")
	}

	if !strings.Contains(e.HTML, "Müller") {
		t.Error("HTML body should contained 'Müller'")
	}
}

func TestParseMultipartWOBoundaryFails(t *testing.T) {
	r := test.OpenTestData("mail", "multipart-wo-boundary.raw")
	_, err := enmime.ReadEnvelope(r)
	if err == nil {
		t.Fatal("Expecting parsing to fail")
	}

	if !strings.Contains(err.Error(), "unable to locate boundary param in Content-Type header") {
		t.Fatal("Expecting for unable to locate boundary error")
	}
}

func TestParseMultipartWOBoundaryAsSinglepart(t *testing.T) {
	r := test.OpenTestData("mail", "multipart-wo-boundary.raw")
	p := enmime.NewParser(enmime.MultipartWOBoundaryAsSinglePart(true))
	e, err := p.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if !bytes.Contains(e.Root.Content, []byte(`I'm  multipart message without boundary`)) {
		t.Fatal("Expecting multipart without boundary to be parsed")
	}
}

func TestEnvelopeGetHeader(t *testing.T) {
	// Test empty header
	e := &enmime.Envelope{}
	want := ""
	got := e.GetHeader("Subject")
	if got != want {
		t.Errorf("Subject was: %q, want: %q", got, want)
	}

	// Even non-MIME messages should support encoded-words in headers
	// Also, encoded addresses should be suppored
	r := test.OpenTestData("mail", "qp-ascii-header.raw")
	e, err := enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}

	want = "Test QP Subject!"
	got = e.GetHeader("Subject")
	if got != want {
		t.Errorf("Subject was: %q, want: %q", got, want)
	}

	// Test UTF-8 subject line
	r = test.OpenTestData("mail", "qp-utf8-header.raw")
	e, err = enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want = "MIME UTF8 Test \u00a2 More Text"
	got = e.GetHeader("Subject")
	if got != want {
		t.Errorf("Subject was: %q, want: %q", got, want)
	}
}

func TestEnvelopeGetHeaderKeys(t *testing.T) {
	// Test empty header
	e := &enmime.Envelope{}
	got := e.GetHeaderKeys()
	if got != nil {
		t.Errorf("Headers was: %q, want: nil", got)
	}

	// Even non-MIME messages should support encoded-words in headers
	// Also, encoded addresses should be suppored
	r := test.OpenTestData("mail", "qp-ascii-header.raw")
	e, err := enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}

	want := []string{"Date", "From", "Subject", "To", "X-Mailer"}
	got = e.GetHeaderKeys()
	sort.Strings(got)
	test.DiffStrings(t, got, want)

	// Test UTF-8 subject line
	r = test.OpenTestData("mail", "qp-utf8-header.raw")
	e, err = enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want = []string{"Content-Type", "Date", "From", "Message-Id", "Mime-Version", "Sender", "Subject", "To", "User-Agent"}
	got = e.GetHeaderKeys()
	sort.Strings(got)
	test.DiffStrings(t, got, want)
}

func TestEnvelopeGetHeaderValues(t *testing.T) {
	r := test.OpenTestData("mail", "ctype-bug.raw")
	e, err := enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	// test Received headers
	want := []string{
		"by 10.76.55.35 with SMTP id o3csp106612oap; Fri, 10 Jul 2015 13:12:34 -0700 (PDT)",
		"from mail135-10.atl141.mandrillapp.com (mail135-10.atl141.mandrillapp.com. [198.2.135.10]) by mx.google.com with ESMTPS id k184si6630505ywf.180.2015.07.10.13.12.34 for <deepak@redsift.io> (version=TLSv1.2 cipher=ECDHE-RSA-AES128-GCM-SHA256 bits=128/128); Fri, 10 Jul 2015 13:12:34 -0700 (PDT)",
		"from pmta03.mandrill.prod.atl01.rsglab.com (127.0.0.1) by mail135-10.atl141.mandrillapp.com id hk0jj41sau80 for <deepak@redsift.io>; Fri, 10 Jul 2015 20:12:33 +0000 (envelope-from <bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com>)",
		"from [67.214.212.122] by mandrillapp.com id 163e4a0faf244a2da6b0121cc7af1fe9; Fri, 10 Jul 2015 20:12:33 +0000",
	}
	got := e.GetHeaderValues("received")
	diff := deep.Equal(got, want)
	if diff != nil {
		t.Errorf("Got: %+v, want: %+v", got, want)
	}
}

func TestEnvelopeSetHeader(t *testing.T) {
	r := test.OpenTestData("mail", "qp-utf8-header.raw")
	e, err := enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	// replace existing header
	want := "André Pirard <PIRARD@vm1.ulg.ac.be>"
	err = e.SetHeader("To", []string{want})
	if err != nil {
		t.Fatal(err)
	}
	got := e.GetHeader("To")
	if got != want {
		t.Errorf("Got: %q, want: %q", got, want)
	}

	// replace existing header with multiple values
	wantSlice := []string{"Mirosław Marczak <marczak@inbucket.com>", "James Hillyerd <jamehi03@jamehi03lx.noa.com>"}
	err = e.SetHeader("To", wantSlice)
	if err != nil {
		t.Fatal(err)
	}
	gotSlice := e.GetHeaderValues("to")
	diff := deep.Equal(gotSlice, wantSlice)
	if diff != nil {
		t.Errorf("Got: %+v, want: %+v", gotSlice, wantSlice)
	}

	// replace non-existing header
	want = "foobar"
	err = e.SetHeader("X-Foo-Bar", []string{want})
	if err != nil {
		t.Fatal(err)
	}
	got = e.GetHeader("X-Foo-Bar")
	if got != want {
		t.Errorf("Got: %q, want: %q", got, want)
	}
}

func TestEnvelopeAddHeader(t *testing.T) {
	r := test.OpenTestData("mail", "qp-utf8-header.raw")
	e, err := enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	// add to existing header
	to := "James Hillyerd <jamehi03@jamehi03lx.noa.com>"
	wantSlice := []string{"\"Mirosław Marczak\" <marczak@inbucket.com>", "James Hillyerd <jamehi03@jamehi03lx.noa.com>"}
	err = e.AddHeader("To", to)
	if err != nil {
		t.Fatal(err)
	}
	gotSlice := e.GetHeaderValues("To")
	diff := deep.Equal(gotSlice, wantSlice)
	if diff != nil {
		t.Errorf("Got: %+v, want: %+v", gotSlice, wantSlice)
	}

	// add to non-existing header
	want := "foobar"
	err = e.AddHeader("X-Foo-Bar", want)
	if err != nil {
		t.Fatal(err)
	}
	got := e.GetHeader("X-Foo-Bar")
	if got != want {
		t.Errorf("Got: %q, want: %q", got, want)
	}
}

func TestEnvelopeDeleteHeader(t *testing.T) {
	r := test.OpenTestData("mail", "qp-utf8-header.raw")
	e, err := enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	// delete user-agent header
	err = e.DeleteHeader("User-Agent")
	if err != nil {
		t.Fatal(err)
	}
	got := e.GetHeader("User-Agent")
	want := ""
	if got != want {
		t.Errorf("Got: %q, want: %q", got, want)
	}
}

func TestEnvelopeAddressListInDifferentFormats(t *testing.T) {
	template := `Date: Mon, 23 Jun 2015 11:40:36 -0400
From: Gopher <from@example.com>
To: %s
Subject: Gophers at Gophercon

Message body

`
	testCases := []struct {
		input string
		want  []*mail.Address
	}{
		{
			"Example User <user@example.com>",
			[]*mail.Address{{
				Name:    "Example User",
				Address: "user@example.com"}},
		},
		{
			"a@h, b@h, c@h",
			[]*mail.Address{
				{Name: "", Address: "a@h"},
				{Name: "", Address: "b@h"},
				{Name: "", Address: "c@h"},
			},
		},
		{
			"=?utf-8?Q?ze=3AStore=20?= =?utf-8?Q?Orange=20?= =?utf-8?Q?Premium=20?= =?utf-8?Q?Reseller?= <noreply@mail.zestore.ru>",
			[]*mail.Address{{
				Name:    "ze:Store Orange Premium Reseller",
				Address: "noreply@mail.zestore.ru"}},
		},
		{
			`"=?UTF-8?Q?Miros=C5=82aw_Marczak?=" <marczak@inbucket.com>`,
			[]*mail.Address{{
				Name:    "Mirosław Marczak",
				Address: "marczak@inbucket.com",
			}},
		},
		{
			"a@example.com;b@example.com",
			[]*mail.Address{
				{Name: "", Address: "a@example.com"},
				{Name: "", Address: "b@example.com"},
			},
		},
		{
			`"Joe @ Company" <joe@company.com>;<other@company.com>`,
			[]*mail.Address{
				{Name: "Joe @ Company", Address: "joe@company.com"},
				{Name: "", Address: "other@company.com"},
			},
		},
		{
			`=?UTF-8?B?dGVzdG5hbWU=?= <=?UTF-8?B?dGVzdG5hbWU=?=@example.com>`,
			[]*mail.Address{
				{Name: "testname", Address: "testname@example.com"},
			},
		},
		{
			// Address folded onto the next line, issue #387
			"First <first@example.com>, Second <second@example\n .com>,\n Third <third@example.com>",
			[]*mail.Address{
				{Name: "First", Address: "first@example.com"},
				{Name: "Second", Address: "second@example.com"},
				{Name: "Third", Address: "third@example.com"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			r := strings.NewReader(fmt.Sprintf(template, tc.input))
			e, err := enmime.ReadEnvelope(r)
			if err != nil {
				t.Errorf("Failed to read envelope for %q:\n%v", tc.input, err)
				return
			}

			got, err := e.AddressList("to")
			if err != nil {
				t.Errorf("Failed to parse address list %q:\n%v", tc.input, err)
				return
			}

			if len(got) != len(tc.want) {
				t.Errorf("Got %v addresses, wanted %v", len(got), len(tc.want))
				return
			}

			for i, want := range tc.want {
				if got[i].Name != want.Name {
					t.Errorf("Got Name %q, wanted %q", got[i].Name, want.Name)
				}
				if got[i].Address != want.Address {
					t.Errorf("Got Address %q, wanted %q", got[i].Address, want.Address)
				}
			}
		})
	}
}

func TestEnvelopeAddressList(t *testing.T) {
	// Test empty header
	e := &enmime.Envelope{}
	_, err := e.AddressList("To")
	if err == nil {
		t.Error("AddressList(\"To\") should have returned err, got nil")
	}

	r := test.OpenTestData("mail", "qp-utf8-header.raw")
	e, err = enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	_, err = e.AddressList("BCC")
	if err == nil {
		t.Error("AddressList(\"BCC\") should have returned err, got nil")
	}

	_, err = e.AddressList("Subject")
	if err == nil {
		t.Error("AddressList(\"Subject\") should have returned err, got nil")
	}

	toAddresses, err := e.AddressList("To")
	if err != nil {
		t.Fatal("Failed to parse To list:", err)
	}
	if len(toAddresses) != 1 {
		t.Fatalf("len(toAddresses) == %v, want: %v", len(toAddresses), 1)
	}

	// Confirm address name was decoded properly
	want := "Mirosław Marczak"
	got := toAddresses[0].Name
	if got != want {
		t.Errorf("To was: %q, want: %q", got, want)
	}

	senderAddresses, err := e.AddressList("Sender")
	if err != nil {
		t.Fatal("Failed to parse Sender list:", err)
	}
	if len(senderAddresses) != 1 {
		t.Fatalf("len(senderAddresses) == %v, want: %v", len(senderAddresses), 1)
	}

	// Confirm address name was decoded properly
	want = "André Pirard"
	got = senderAddresses[0].Name
	if got != want {
		t.Errorf("Sender was: %q, want: %q", got, want)
	}
}

func TestDetectCharacterSetInHTML(t *testing.T) {
	msg := test.OpenTestData("mail", "non-mime-missing-charset.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}
	if strings.ContainsRune(e.HTML, 0x80) {
		t.Error("HTML body should not have contained a Windows CP1250 Euro Symbol")
	}
	if !strings.ContainsRune(e.HTML, 0x20ac) {
		t.Error("HTML body should have contained a Unicode Euro Symbol")
	}
}

func TestAttachmentOnly(t *testing.T) {
	var aTests = []struct {
		filename       string
		attachmentsLen int
		inlinesLen     int
	}{
		{filename: "attachment-only.raw", attachmentsLen: 1, inlinesLen: 0},
		{filename: "attachment-only-inline.raw", attachmentsLen: 0, inlinesLen: 1},
	}

	for _, a := range aTests {
		// Mail with disposition attachment
		msg := test.OpenTestData("mail", a.filename)
		e, err := enmime.ReadEnvelope(msg)
		if err != nil {
			t.Fatal("Failed to parse MIME:", err)
		}
		if len(e.Attachments) != a.attachmentsLen {
			t.Fatal("len(Attachments) got:", len(e.Attachments), "want:", a.attachmentsLen)
		}
		if a.attachmentsLen > 0 {
			got := e.Attachments[0].Content
			if !bytes.HasPrefix(got, []byte{0x89, 'P', 'N', 'G'}) {
				t.Errorf("Content should be PNG image, got: %v", got)
			}
		}
		if len(e.Inlines) != a.inlinesLen {
			t.Fatal("len(Inlines) got:", len(e.Inlines), "want:", a.inlinesLen)
		}
		if a.inlinesLen > 0 {
			got := e.Inlines[0].Content
			if !bytes.HasPrefix(got, []byte{0x89, 'P', 'N', 'G'}) {
				t.Errorf("Content should be PNG image, got: %v", got)
			}
		}
		// Check, if root header is set
		if len(e.Root.Header) < 1 {
			t.Error("No root header defined, but must be set from binary only part.")
		}
		// Check, that the root part has content
		if len(e.Root.Content) == 0 {
			t.Error("Root part of envelope has no content.")
		}
	}
}

func TestDuplicateParamsInMime(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-duplicate-param.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if e.Attachments[0].FileName != "Invoice_302232133150612.pdf" {
		t.Fatal("Mail should have a part with filename Invoice_302232133150612.pdf")
	}
}

func TestUnquotedSpecialCharParamsInMime(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-unquoted-tspecials-param.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if e.Attachments[0].FileName != "Invoice_(302232133150612).pdf" {
		t.Fatal("Mail should have a part with filename Invoice_(302232133150612).pdf")
	}
}

func TestBadAddressHeaderInMime(t *testing.T) {
	msg := test.OpenTestData("mail", "malformed-multiple-address-header-values.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	froms, err := e.AddressList("From")
	if err != nil {
		t.Log(err)
	}

	if len(froms) < 1 {
		t.Fatal("From header should have at least one entry")
	}
}

func TestBadContentTypeInMime(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-bad-content-type.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if e.Attachments[0].FileName != "Invoice_302232133150612.pdf" {
		t.Fatal("Mail should have a part with filename Invoice_302232133150612.pdf")
	}
}

func TestBadContentTransferEncodingInMime(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-bad-content-transfer-encoding.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	var expectedErrorPresent bool
	for _, v := range e.Errors {
		if v.Name == enmime.ErrorMalformedBase64 && v.Severe {
			expectedErrorPresent = true
		}
	}
	if !expectedErrorPresent {
		t.Fatal("Mail should have a severe malformed base64 error")
	}
}

func TestBadMime8bitFilename(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-bad-8bit-filename.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if strings.TrimSpace(e.Text) != "Text part" {
		t.Fatal("Text part not parsed correctly")
	}
	if len(e.Attachments) != 1 {
		t.Fatal("Wrong number of attachments")
	}
	if e.Attachments[0].FileName != "管理.doc" {
		t.Fatal("Wrong attachment name")
	}
}

func TestBlankMediaName(t *testing.T) {
	msg := test.OpenTestData("mail", "mime-blank-media-name.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if e.Attachments[0].FileName != "Invoice_302232133150612.pdf" {
		t.Fatal("Mail should have a part with filename Invoice_302232133150612.pdf")
	}
}

func TestEnvelopeHeaders(t *testing.T) {
	headers := map[string]string{
		"Received-Spf":           "pass (google.com: domain of bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com designates 198.2.135.10 as permitted sender) client-ip=198.2.135.10;",
		"To":                     "<deepak@redsift.io>",
		"Domainkey-Signature":    "a=rsa-sha1; c=nofws; q=dns; s=mandrill; d=papertrailapp.com; b=Cv9EE+3+CO+puDhpfQOsuwuP6YqJQBA/Z6OofPTXqWf/Asr/edsi7aoXIE+forQ/q8DjhhMMuMiD bQ1tlRXMFckw08GjqU7RN+ouwJEMXOpzxUgp6OwrITvddwhddEg6H3uYRva5pNJqonDDykshHyjA EVeAdcY4tjYQrcRxw/0=;",
		"Dkim-Signature":         "v=1; a=rsa-sha1; c=relaxed/relaxed; s=mandrill; d=papertrailapp.com; h=From:Subject:To:Message-Id:Date:MIME-Version:Content-Type; i=support@papertrailapp.com; bh=2tw/BU7QN7gmFr2K2wnVpETYxbU=; b=T+PzWzjbOoKO3jNANsmqsnbM+gnbgT9EQBP8DOSno75iHQ9AuU6xcDCPctvJt50Exr6aTs9qJmEG baCa39danDRIx5zXsdaSy34+SKfDODdgmwEEfKFeULQGPwF1g73tXeX4k0kwt+bm6f0baWbaLwR1 RdhUd42jEMossTKuD9w= v=1; a=rsa-sha256; c=relaxed/relaxed; d=mandrillapp.com; i=@mandrillapp.com; q=dns/txt; s=mandrill; t=1436559153; h=From : Subject : To : Message-Id : Date : MIME-Version : Content-Type : From : Subject : Date : X-Mandrill-User : List-Unsubscribe; bh=eW2QM8XcfLCwIBTvTJaT619pYOD3YrxBvxC9cZ2gxe0=; b=quxFFNbO04KKNNB8yMd9Zch6wogobVbNFlpGIOQI/jA9FuhdZvMxQwwZ2jeno7c17v2eXY Vp3c1vwvVERCboNaPwwxrKkrhqMxM8rb15n8xM3v0IplkQ3vs9G5agiTT1qqxErsrS6xAqmj UNUPKEXuSjr24HqmQzxPry0aIgHdI=",
		"Message-Id":             "<55a02731af510_7b0b33f2c7821d@pt02w01.papertrailapp.com.tmail>",
		"X-Report-Abuse":         "Please forward a copy of this message, including all headers, to abuse@mandrill.com You can also report abuse here: http://mandrillapp.com/contact/abuse?id=30112948.163e4a0faf244a2da6b0121cc7af1fe9",
		"Mime-Version":           "1.0",
		"Return-Path":            "<bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com> <bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com>",
		"Authentication-Results": "mx.google.com; spf=pass (google.com: domain of bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com designates 198.2.135.10 as permitted sender) smtp.mail=bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com; dkim=pass header.i=@papertrailapp.com; dkim=pass header.i=@mandrillapp.com",
		"From":                   "Papertrail <support@papertrailapp.com>",
		"Subject":                "Welcome to Papertrail",
		"Content-Type":           `multipart/alternative; boundary="_av-rPFkvS5QROAYLq2cQTUr1w"`,
		"X-Mandrill-User":        "md_30112948",
		"Delivered-To":           "deepak@redsift.io",
		"Received":               "by 10.76.55.35 with SMTP id o3csp106612oap; Fri, 10 Jul 2015 13:12:34 -0700 (PDT) from mail135-10.atl141.mandrillapp.com (mail135-10.atl141.mandrillapp.com. [198.2.135.10]) by mx.google.com with ESMTPS id k184si6630505ywf.180.2015.07.10.13.12.34 for <deepak@redsift.io> (version=TLSv1.2 cipher=ECDHE-RSA-AES128-GCM-SHA256 bits=128/128); Fri, 10 Jul 2015 13:12:34 -0700 (PDT) from pmta03.mandrill.prod.atl01.rsglab.com (127.0.0.1) by mail135-10.atl141.mandrillapp.com id hk0jj41sau80 for <deepak@redsift.io>; Fri, 10 Jul 2015 20:12:33 +0000 (envelope-from <bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com>) from [67.214.212.122] by mandrillapp.com id 163e4a0faf244a2da6b0121cc7af1fe9; Fri, 10 Jul 2015 20:12:33 +0000",
		"X-Received":             "by 10.170.119.147 with SMTP id l141mr25507408ykb.89.1436559154116; Fri, 10 Jul 2015 13:12:34 -0700 (PDT)",
		"Date":                   "Fri, 10 Jul 2015 20:12:33 +0000",
	}

	msg := test.OpenTestData("mail", "ctype-bug.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if len(e.Root.Header) != len(headers) {
		t.Errorf("Failed to extract expected headers. Got %v headers, expected %v",
			len(e.Root.Header), len(headers))
	}

	for k := range headers {
		if e.Root.Header[k] == nil {
			t.Errorf("Header named %q was missing, want it to exist", k)
		}
	}

	for k, v := range e.Root.Header {
		if _, ok := headers[k]; !ok {
			t.Errorf("Got header named %q, did not expect it to exist", k)
			continue
		}
		for _, val := range v {
			if !strings.Contains(headers[k], val) {
				t.Errorf("Got header %q with value %q, wanted value contained in:\n%q",
					k, val, headers[k])
			}
		}
	}
}

func TestInlineTextBody(t *testing.T) {
	headers := map[string]string{
		"To":                        "<info@stuff.com>",
		"Message-Id":                "<E1cbNWz-0002jP-2E@stuf.com>",
		"Mime-Version":              "1.0",
		"From":                      "Chris Garrett <cgarrett@stuff.com>",
		"Subject":                   "Text body only with disposition inline",
		"Content-Type":              `text/html; charset="UTF-8"`,
		"Content-Disposition":       "inline",
		"Content-Transfer-Encoding": "quoted-printable",
		"Date":                      "Wed, 8 Feb 2017 03:23:13 -0500",
	}

	msg := test.OpenTestData("mail", "attachment-only-inline-quoted-printable.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "Just some html content"
	if e.Text != want {
		t.Errorf("Got Text with value %s, wanted value:\n%s", e.Text, want)
	}
	if !strings.Contains(e.HTML, want) {
		t.Errorf("Expected %q to contain %q", e.HTML, want)
	}

	if len(e.Root.Header) != len(headers) {
		t.Errorf("Failed to extract expected headers. Got %v headers, expected %v",
			len(e.Root.Header), len(headers))
	}

	for k := range headers {
		if e.Root.Header[k] == nil {
			t.Errorf("Header named %q was missing, want it to exist", k)
		}
	}

	for k, v := range e.Root.Header {
		if _, ok := headers[k]; !ok {
			t.Errorf("Got header named %q, did not expect it to exist", k)
			continue
		}
		for _, val := range v {
			if !strings.Contains(headers[k], val) {
				t.Errorf("Got header %q with value %q, wanted value contained in:\n%q",
					k, val, headers[k])
			}
		}
	}
}

func TestBinaryOnlyBodyHeaders(t *testing.T) {
	headers := map[string]string{
		"To":                        "bob@test.com",
		"From":                      "alice@test.com",
		"Subject":                   "Test",
		"Message-Id":                "<56A0AA5F.4020203@test.com>",
		"Date":                      "Thu, 21 Jan 2016 10:52:31 +0100",
		"Mime-Version":              "1.0",
		"Content-Type":              `image/jpeg; name="favicon.jpg"`,
		"Content-Transfer-Encoding": "base64",
		"Content-Disposition":       `attachment; filename="favicon.jpg"`,
	}

	msg := test.OpenTestData("mail", "attachment-only.raw")
	e, err := enmime.ReadEnvelope(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if len(e.Root.Header) != len(headers) {
		t.Errorf("Failed to extract expected headers. Got %v headers, expected %v",
			len(e.Root.Header), len(headers))
	}

	for k := range headers {
		if e.Root.Header[k] == nil {
			t.Errorf("Header named %q was missing, want it to exist", k)
		}
	}

	for k, v := range e.Root.Header {
		if _, ok := headers[k]; !ok {
			t.Errorf("Got header named %q, did not expect it to exist", k)
			continue
		}
		for _, val := range v {
			if !strings.Contains(headers[k], val) {
				t.Errorf("Got header %q with value %q, wanted value contained in:\n%q",
					k, val, headers[k])
			}
		}
	}
}

func TestEnvelopeEpilogue(t *testing.T) {
	msg := test.OpenTestData("mail", "epilogue-sample.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	got := string(e.Root.Epilogue)
	want := "Potentially malicious content\n"
	if got != want {
		t.Errorf("Epilogue == %q, want: %q", got, want)
	}
}

func TestCloneEnvelope(t *testing.T) {
	msg := test.OpenTestData("mail", "other-multi-related.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	clone := e.Clone()
	test.CompareEnvelope(t, clone, e)
}

func TestNilHeaderValues(t *testing.T) {
	e := &enmime.Envelope{}
	values := e.GetHeaderValues("")
	if len(values) > 0 {
		t.Fatal("No header values should be available, failed")
	}
}

func TestSetHeaderEmptyName(t *testing.T) {
	e := &enmime.Envelope{}
	err := e.SetHeader("", nil)
	if err == nil || err.Error() != "provide non-empty header name" {
		t.Fatal("Cannot set a header without a name, failed")
	}
}

func TestAddHeaderEmptyName(t *testing.T) {
	e := &enmime.Envelope{}
	err := e.AddHeader("", "")
	if err == nil || err.Error() != "provide non-empty header name" {
		t.Fatal("Cannot add a header without a name, failed")
	}
}

func TestDeleteHeaderEmptyName(t *testing.T) {
	e := &enmime.Envelope{}
	err := e.DeleteHeader("")
	if err == nil || err.Error() != "provide non-empty header name" {
		t.Fatal("Cannot add a header without a name, failed")
	}
}

func TestNilEnvelopeClone(t *testing.T) {
	var e *enmime.Envelope
	if e.Clone() != nil {
		t.Fatal("Clone of nil envelope is not nil, failed")
	}
}

func TestEnvelopeSend(t *testing.T) {
	sender := &mockSender{}
	msg := test.OpenTestData("mail", "attachment.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	from := "from@example.com"
	to := "to@example.com"
	if err := e.SetHeader("From", []string{"Sender Name <" + from + ">"}); err != nil {
		t.Fatal(err)
	}
	if err := e.SetHeader("To", []string{"Recipient Name <" + to + ">"}); err != nil {
		t.Fatal(err)
	}

	err = e.Send(sender)
	if err != nil {
		t.Fatal(err)
	}

	if sender.from != from {
		t.Errorf("Got reverse-path %q, wanted %q", sender.from, from)
	}
	test.DiffStrings(t, sender.to, []string{to})
	if !bytes.Contains(sender.msg, []byte("A text section")) {
		t.Errorf("msg bytes did not contain text body %q", "A text section")
	}
	if !bytes.Contains(sender.msg, []byte("test.html")) {
		t.Errorf("msg bytes did not contain attachment %q", "test.html")
	}
}

func TestEnvelopeSendWithReversePath(t *testing.T) {
	sender := &mockSender{}
	msg := test.OpenTestData("mail", "attachment.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if err := e.SetHeader("To", []string{"to@example.com"}); err != nil {
		t.Fatal(err)
	}

	const ret = "return@example.com"
	err = e.SendWithReversePath(sender, ret)
	if err != nil {
		t.Fatal(err)
	}

	if sender.from != ret {
		// The From header should not be provided to Sender.Send().
		t.Errorf("Got reverse-path %q, wanted %q", sender.from, ret)
	}
	test.DiffStrings(t, sender.to, []string{"to@example.com"})
}

func TestEnvelopeSendMultipleHeaderValues(t *testing.T) {
	sender := &mockSender{}
	msg := strings.NewReader(`From: from@example.com
To: one@example.com
To: two@example.com
Cc: three@example.com
Subject: Folded

Body
`)
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	err = e.Send(sender)
	if err != nil {
		t.Fatal(err)
	}

	if sender.from != "from@example.com" {
		t.Errorf("Got reverse-path %q, wanted %q", sender.from, "from@example.com")
	}
	// A second "To:" value must not be silently dropped.
	test.DiffStrings(t, sender.to,
		[]string{"one@example.com", "two@example.com", "three@example.com"})
}

func TestEnvelopeSendDuplicateRecipients(t *testing.T) {
	sender := &mockSender{}
	msg := strings.NewReader(`From: from@example.com
To: dup@example.com, dup@example.com
Cc: other@example.com
Subject: Dupes

Body
`)
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	err = e.Send(sender)
	if err != nil {
		t.Fatal(err)
	}

	test.DiffStrings(t, sender.to, []string{"dup@example.com", "other@example.com"})
}

func TestEnvelopeSendNoRecipients(t *testing.T) {
	sender := &mockSender{}
	msg := strings.NewReader(`From: from@example.com
Subject: Nobody

Body
`)
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	err = e.Send(sender)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != enmime.ErrorMissingRecipient {
		t.Fatalf("Unexpected error, wanted %q got %s", enmime.ErrorMissingRecipient, err)
	}
}

func TestEnvelopeSendMissingFrom(t *testing.T) {
	sender := &mockSender{}
	msg := strings.NewReader(`To: to@example.com
Subject: No From

Body
`)
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	err = e.Send(sender)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "From") {
		t.Errorf("Error %q should mention the From header", err)
	}
}

func TestEnvelopeSendUnparsableFrom(t *testing.T) {
	sender := &mockSender{}
	msg := strings.NewReader(`From: not-an-address
To: to@example.com
Subject: Bad From

Body
`)
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	err = e.Send(sender)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestEnvelopeSendUnparsableRecipient(t *testing.T) {
	sender := &mockSender{}
	msg := strings.NewReader(`From: from@example.com
To: not-an-address
Subject: Bad recipient

Body
`)
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	err = e.Send(sender)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestEnvelopeSendNilEnvelope(t *testing.T) {
	var e *enmime.Envelope
	if err := e.Send(&mockSender{}); err == nil {
		t.Fatal("Expected error sending nil Envelope, got nil")
	}
}

func TestEnvelopeSendBcc(t *testing.T) {
	sender := &mockSender{}
	msg := strings.NewReader(`From: from@example.com
To: to@example.com
Bcc: secret@example.com
Subject: Bcc

Body
`)
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	err = e.Send(sender)
	if err != nil {
		t.Fatal(err)
	}

	test.DiffStrings(t, sender.to, []string{"to@example.com", "secret@example.com"})
	if bytes.Contains(sender.msg, []byte("Bcc:")) {
		t.Error("Transmitted message must not contain a Bcc header, see RFC 5322 section 3.6.3")
	}
	if got := e.GetHeader("Bcc"); got != "secret@example.com" {
		t.Errorf("Envelope Bcc header == %q, want it restored to %q", got, "secret@example.com")
	}
}

func TestEnvelopeSendRoundTripGolden(t *testing.T) {
	sender := &mockSender{}
	msg := test.OpenTestData("mail", "mime-mixed.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	err = e.Send(sender)
	if err != nil {
		t.Fatal(err)
	}

	// The parsed message should survive re-encoding, boundaries included.
	reparsed, err := enmime.ReadEnvelope(bytes.NewReader(sender.msg))
	if err != nil {
		t.Fatal("Failed to reparse sent message:", err)
	}
	if reparsed.Text != e.Text {
		t.Errorf("Text changed on round trip: %q != %q", reparsed.Text, e.Text)
	}
	if reparsed.HTML != e.HTML {
		t.Errorf("HTML changed on round trip: %q != %q", reparsed.HTML, e.HTML)
	}
	if len(reparsed.Attachments) != len(e.Attachments) {
		t.Errorf("Attachment count changed: %d != %d",
			len(reparsed.Attachments), len(e.Attachments))
	}

	test.DiffGolden(t, sender.msg, "testdata", "encode", "envelope-send-mime-mixed.golden")
}

func TestEnvelopeSendRejectsCRLFAddress(t *testing.T) {
	// ParseAddressList decodes RFC 2047 encoded words, so these headers can decode to addresses
	// embedding CRLF; they must be rejected rather than handed to the Sender.
	crlf := "=?utf-8?q?a=0D=0Ab?=@example.com"

	for _, tc := range []struct {
		name string
		msg  string
		ret  string
	}{
		{"recipient", "From: from@example.com\nTo: " + crlf + "\nSubject: x\n\nBody\n", ""},
		{"reverse-path", "From: " + crlf + "\nTo: to@example.com\nSubject: x\n\nBody\n", ""},
		{"explicit reverse-path", "From: from@example.com\nTo: to@example.com\nSubject: x\n\nBody\n", "bad\r\naddress"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sender := &mockSender{}
			e, err := enmime.ReadEnvelope(strings.NewReader(tc.msg))
			if err != nil {
				t.Fatal("Failed to parse MIME:", err)
			}

			if tc.ret == "" {
				err = e.Send(sender)
			} else {
				err = e.SendWithReversePath(sender, tc.ret)
			}
			if err == nil {
				t.Fatalf("Expected error, got nil; sender got %q -> %q", sender.from, sender.to)
			}
			if sender.msg != nil {
				t.Error("Sender must not be invoked with an invalid address")
			}
		})
	}
}

func TestEnvelopeSetAddressHeader(t *testing.T) {
	e, err := enmime.ReadEnvelope(strings.NewReader("From: old@example.com\nTo: old@example.com\nSubject: x\n\nBody\n"))
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	addrs := []*mail.Address{
		{Name: "Bob", Address: "bob@example.com"},
		{Name: "José García", Address: "jose@example.com"},
		{Name: "", Address: "plain@example.com"},
		{Name: "Ünïcödé with (parens); semis", Address: "u@example.com"},
	}
	if err := e.SetAddressHeader("To", addrs); err != nil {
		t.Fatal(err)
	}

	// All previous values are replaced by the single formatted one.
	if values := e.GetHeaderValues("To"); len(values) != 1 {
		t.Fatalf("Got %d To values, want 1", len(values))
	}

	// A non-ASCII display name is written as an encoded-word ahead of its address, never as one
	// encoded-word covering the whole list.
	raw := e.Root.Header["To"][0]
	if !strings.Contains(raw, "=?utf-8?") {
		t.Errorf("To header %q does not contain an encoded-word", raw)
	}
	if strings.Contains(raw, "=? ") {
		t.Errorf("To header %q looks like an encoded-word wrapping the address list", raw)
	}
	for _, addr := range addrs {
		if !strings.Contains(raw, "<"+addr.Address+">") {
			t.Errorf("To header %q is missing address %s", raw, addr.Address)
		}
	}

	// Round trip: the stored value must parse back into the addresses we set.
	got, err := enmime.ParseAddressList(raw)
	if err != nil {
		t.Fatalf("ParseAddressList(%q): %v", raw, err)
	}
	if len(got) != len(addrs) {
		t.Fatalf("Got %d addresses, want %d", len(got), len(addrs))
	}
	for i, want := range addrs {
		if got[i].Name != want.Name || got[i].Address != want.Address {
			t.Errorf("Address %d = %q <%s>, want %q <%s>",
				i, got[i].Name, got[i].Address, want.Name, want.Address)
		}
	}

	// AddressList sees the same thing.
	viaEnv, err := e.AddressList("To")
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(viaEnv, got); diff != nil {
		t.Errorf("AddressList disagrees with ParseAddressList: %v", diff)
	}
}

func TestEnvelopeSetAddressHeaderASCIIName(t *testing.T) {
	e, err := enmime.ReadEnvelope(strings.NewReader("From: old@example.com\n\nBody\n"))
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	// A plain ASCII name is left alone rather than encoded or quoted.
	if err := e.SetAddressHeader("From", []*mail.Address{
		{Name: "Bob", Address: "bob@example.com"},
	}); err != nil {
		t.Fatal(err)
	}
	if got := e.Root.Header["From"][0]; got != "Bob <bob@example.com>" {
		t.Errorf("Got %q, want %q", got, "Bob <bob@example.com>")
	}
}

func TestEnvelopeSetAddressHeaderPreEncodedName(t *testing.T) {
	e, err := enmime.ReadEnvelope(strings.NewReader("From: old@example.com\n\nBody\n"))
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	// An already-encoded ASCII name is emitted as an encoded-word, not as a quoted-string.  Go's
	// mail.Address.String would quote it, because = and ? are atom characters; the name must stay
	// unquoted for parsers to see the encoded-word.
	if err := e.SetAddressHeader("From", []*mail.Address{
		{Name: "=?utf-8?Q?Jane_Doe?=", Address: "jane@example.com"},
	}); err != nil {
		t.Fatal(err)
	}
	if got := e.Root.Header["From"][0]; got != "=?utf-8?Q?Jane_Doe?= <jane@example.com>" {
		t.Errorf("Got %q, want %q", got, "=?utf-8?Q?Jane_Doe?= <jane@example.com>")
	}

	addrs, err := enmime.ParseAddressList(e.Root.Header["From"][0])
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 || addrs[0].Name != "Jane Doe" || addrs[0].Address != "jane@example.com" {
		t.Errorf("Got %+v, want Jane Doe <jane@example.com>", addrs)
	}
}

func TestEnvelopeSetAddressHeaderRejectsNonAddressHeader(t *testing.T) {
	e, err := enmime.ReadEnvelope(strings.NewReader("From: old@example.com\n\nBody\n"))
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	err = e.SetAddressHeader("Subject", []*mail.Address{{Address: "bob@example.com"}})
	if err == nil {
		t.Fatal("Subject is not an address header, expected error")
	}
	if !strings.Contains(err.Error(), "not an address header") {
		t.Errorf("Got error %q, want it to name the valid address headers", err)
	}

	err = e.SetAddressHeader("", []*mail.Address{{Address: "bob@example.com"}})
	if err == nil || err.Error() != "provide non-empty header name" {
		t.Errorf("Got error %v, want %q", err, "provide non-empty header name")
	}
}

func TestEnvelopeSetAddressHeaderNil(t *testing.T) {
	e, err := enmime.ReadEnvelope(strings.NewReader("From: old@example.com\nTo: first@example.com\n\nBody\n"))
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	// Nil and empty slices clear the header instead of writing an empty value.
	for _, addrs := range [][]*mail.Address{nil, {}, {nil}} {
		if err := e.SetAddressHeader("To", addrs); err != nil {
			t.Fatal(err)
		}
		if values := e.GetHeaderValues("To"); len(values) != 0 {
			t.Errorf("To = %q, want it deleted", values)
		}
	}
}

func TestEnvelopeSetAddressHeaderSendRoundTrip(t *testing.T) {
	sender := &mockSender{}
	msg := test.OpenTestData("mail", "mime-mixed.raw")
	e, err := enmime.ReadEnvelope(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	// The point of the feature: edit an address header with a non-ASCII display name, then send the
	// parsed message without rebuilding it.
	from := []*mail.Address{{Name: "José García", Address: "jose@example.com"}}
	if err := e.SetAddressHeader("From", from); err != nil {
		t.Fatal(err)
	}
	if err := e.Send(sender); err != nil {
		t.Fatal(err)
	}

	if sender.from != "jose@example.com" {
		t.Errorf("Got reverse-path %q, want %q", sender.from, "jose@example.com")
	}

	reparsed, err := enmime.ReadEnvelope(bytes.NewReader(sender.msg))
	if err != nil {
		t.Fatal("Failed to reparse sent message:", err)
	}
	addrs, err := reparsed.AddressList("From")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 {
		t.Fatalf("Got %d From addresses, want 1", len(addrs))
	}
	if addrs[0].Name != "José García" || addrs[0].Address != "jose@example.com" {
		t.Errorf("Got From %q <%s>, want %q <%s>",
			addrs[0].Name, addrs[0].Address, from[0].Name, from[0].Address)
	}

	test.DiffGolden(t, sender.msg, "testdata", "encode", "envelope-set-address-header-send.golden")
}
