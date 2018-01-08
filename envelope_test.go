package enmime_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/internal/test"
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
		t.Errorf("Expected a root part")
	}
	if len(e.Root.Header) != 7 {
		t.Errorf("Expected 7 headers, got %d", len(e.Root.Header))
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
		t.Fatal("Should have no attachments, got:", len(e.Attachments))
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

func TestEnvelopeAddressList(t *testing.T) {
	// Test empty header
	e := &enmime.Envelope{}
	_, err := e.AddressList("To")
	if err == nil {
		t.Error("AddressList(\"Subject\") should have returned err, got nil")
	}

	r := test.OpenTestData("mail", "qp-utf8-header.raw")
	e, err = enmime.ReadEnvelope(r)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
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
			t.Errorf("No root header defined, but must be set from binary only part.")
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
		"Received-Spf": "pass (google.com: domain of bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com designates 198.2.135.10 as permitted sender) client-ip=198.2.135.10;",
		"To":           "<deepak@redsift.io>",
		"Domainkey-Signature":    "a=rsa-sha1; c=nofws; q=dns; s=mandrill; d=papertrailapp.com; b=Cv9EE+3+CO+puDhpfQOsuwuP6YqJQBA/Z6OofPTXqWf/Asr/edsi7aoXIE+forQ/q8DjhhMMuMiD bQ1tlRXMFckw08GjqU7RN+ouwJEMXOpzxUgp6OwrITvddwhddEg6H3uYRva5pNJqonDDykshHyjA EVeAdcY4tjYQrcRxw/0=;",
		"Dkim-Signature":         "v=1; a=rsa-sha1; c=relaxed/relaxed; s=mandrill; d=papertrailapp.com; h=From:Subject:To:Message-Id:Date:MIME-Version:Content-Type; i=support@papertrailapp.com; bh=2tw/BU7QN7gmFr2K2wnVpETYxbU=; b=T+PzWzjbOoKO3jNANsmqsnbM+gnbgT9EQBP8DOSno75iHQ9AuU6xcDCPctvJt50Exr6aTs9qJmEG baCa39danDRIx5zXsdaSy34+SKfDODdgmwEEfKFeULQGPwF1g73tXeX4k0kwt+bm6f0baWbaLwR1 RdhUd42jEMossTKuD9w= v=1; a=rsa-sha256; c=relaxed/relaxed; d=mandrillapp.com; i=@mandrillapp.com; q=dns/txt; s=mandrill; t=1436559153; h=From : Subject : To : Message-Id : Date : MIME-Version : Content-Type : From : Subject : Date : X-Mandrill-User : List-Unsubscribe; bh=eW2QM8XcfLCwIBTvTJaT619pYOD3YrxBvxC9cZ2gxe0=; b=quxFFNbO04KKNNB8yMd9Zch6wogobVbNFlpGIOQI/jA9FuhdZvMxQwwZ2jeno7c17v2eXY Vp3c1vwvVERCboNaPwwxrKkrhqMxM8rb15n8xM3v0IplkQ3vs9G5agiTT1qqxErsrS6xAqmj UNUPKEXuSjr24HqmQzxPry0aIgHdI=",
		"Message-Id":             "<55a02731af510_7b0b33f2c7821d@pt02w01.papertrailapp.com.tmail>",
		"X-Report-Abuse":         "Please forward a copy of this message, including all headers, to abuse@mandrill.com You can also report abuse here: http://mandrillapp.com/contact/abuse?id=30112948.163e4a0faf244a2da6b0121cc7af1fe9",
		"Mime-Version":           "1.0",
		"Return-Path":            "<bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com> <bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com>",
		"Authentication-Results": "mx.google.com; spf=pass (google.com: domain of bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com designates 198.2.135.10 as permitted sender) smtp.mail=bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com; dkim=pass header.i=@papertrailapp.com; dkim=pass header.i=@mandrillapp.com",
		"From":            "Papertrail <support@papertrailapp.com>",
		"Subject":         "Welcome to Papertrail",
		"Content-Type":    `multipart/alternative; boundary="_av-rPFkvS5QROAYLq2cQTUr1w"`,
		"X-Mandrill-User": "md_30112948",
		"Delivered-To":    "deepak@redsift.io",
		"Received":        "by 10.76.55.35 with SMTP id o3csp106612oap; Fri, 10 Jul 2015 13:12:34 -0700 (PDT) from mail135-10.atl141.mandrillapp.com (mail135-10.atl141.mandrillapp.com. [198.2.135.10]) by mx.google.com with ESMTPS id k184si6630505ywf.180.2015.07.10.13.12.34 for <deepak@redsift.io> (version=TLSv1.2 cipher=ECDHE-RSA-AES128-GCM-SHA256 bits=128/128); Fri, 10 Jul 2015 13:12:34 -0700 (PDT) from pmta03.mandrill.prod.atl01.rsglab.com (127.0.0.1) by mail135-10.atl141.mandrillapp.com id hk0jj41sau80 for <deepak@redsift.io>; Fri, 10 Jul 2015 20:12:33 +0000 (envelope-from <bounce-md_30112948.55a02731.v1-163e4a0faf244a2da6b0121cc7af1fe9@mandrill.papertrailapp.com>) from [67.214.212.122] by mandrillapp.com id 163e4a0faf244a2da6b0121cc7af1fe9; Fri, 10 Jul 2015 20:12:33 +0000",
		"X-Received":      "by 10.170.119.147 with SMTP id l141mr25507408ykb.89.1436559154116; Fri, 10 Jul 2015 13:12:34 -0700 (PDT)",
		"Date":            "Fri, 10 Jul 2015 20:12:33 +0000",
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
		"Date": "Wed, 8 Feb 2017 03:23:13 -0500",
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
