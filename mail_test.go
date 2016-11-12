package enmime

import (
	"bufio"
	"bytes"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIdentifySinglePart(t *testing.T) {
	msg := readMessage("non-mime.raw")
	if IsMultipartMessage(msg) {
		t.Error("Failed to identify non-multipart message")
	}
}

func TestIdentifyMultiPart(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	if !IsMultipartMessage(msg) {
		t.Error("Failed to identify multipart MIME message")
	}
}

func TestIdentifyUnknownMultiPart(t *testing.T) {
	msg := readMessage("unknown-part-type.raw")
	if !IsMultipartMessage(msg) {
		t.Error("Failed to identify multipart MIME message of unknown type")
	}
}

func TestParseNonMime(t *testing.T) {
	want := "This is a test mailing"
	msg := readMessage("non-mime.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}
	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}
	if !strings.Contains(mime.Text, want) {
		t.Errorf("Expected %q to contain %q", mime.Text, want)
	}
	if mime.HTML != "" {
		t.Errorf("Expected no HTML body, got %q", mime.HTML)
	}
}

func TestParseNonMimeHTML(t *testing.T) {
	msg := readMessage("non-mime-html.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}
	if !mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be true")
	}

	want := "This is *a* *test* mailing"
	if !strings.Contains(mime.Text, want) {
		t.Errorf("Expected %q to contain %q", mime.Text, want)
	}

	want = "<span>This</span>"
	if !strings.Contains(mime.HTML, want) {
		t.Errorf("Expected %q to contain %q", mime.HTML, want)
	}
}

func TestParseMimeTree(t *testing.T) {
	msg := readMessage("attachment.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}
	if mime.Root == nil {
		t.Error("Message should have a root node")
	}
}

func TestParseInlineText(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}

	want := "Test of text section"
	if mime.Text != want {
		t.Error("got:", mime.Text, "want:", want)
	}
}

func TestParseMultiMixedText(t *testing.T) {
	msg := readMessage("mime-mixed.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}

	want := "Section one\n\n--\nSection two"
	if mime.Text != want {
		t.Error("Text parts should concatenate, got:", mime.Text, "want:", want)
	}
}

func TestParseMultiSignedText(t *testing.T) {
	msg := readMessage("mime-signed.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}

	want := "Section one\n\n--\nSection two"
	if mime.Text != want {
		t.Error("Text parts should concatenate, got:", mime.Text, "want:", want)
	}
}

func TestParseQuotedPrintable(t *testing.T) {
	msg := readMessage("quoted-printable.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}

	want := "Phasellus sit amet arcu"
	if !strings.Contains(mime.Text, want) {
		t.Errorf("Text: %q should contain: %q", mime.Text, want)
	}
}

func TestParseQuotedPrintableMime(t *testing.T) {
	msg := readMessage("quoted-printable-mime.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}

	want := "Nullam venenatis ante"
	if !strings.Contains(mime.Text, want) {
		t.Errorf("Text: %q should contain: %q", mime.Text, want)
	}
}

func TestParseInlineHTML(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "<html>"
	if !strings.Contains(mime.HTML, want) {
		t.Errorf("HTML: %q should contain: %q", mime.Text, want)
	}

	want = "Test of HTML section"
	if !strings.Contains(mime.HTML, want) {
		t.Errorf("HTML: %q should contain: %q", mime.Text, want)
	}
}

func TestParseAttachment(t *testing.T) {
	msg := readMessage("attachment.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}

	want := "A text section"
	if !strings.Contains(mime.Text, want) {
		t.Errorf("Text: %q should contain: %q", mime.Text, want)
	}
	if mime.HTML != "" {
		t.Error("mime.HTML should be empty, attachment is not for display, got:", mime.HTML)
	}
	if len(mime.Inlines) > 0 {
		t.Error("Should have no inlines, got:", len(mime.Inlines))
	}
	if len(mime.Attachments) != 1 {
		t.Fatal("Should have a single attachment, got:", len(mime.Attachments))
	}

	want = "test.html"
	got := mime.Attachments[0].FileName()
	if got != want {
		t.Error("FileName() got:", got, "want:", want)
	}

	want = "<html>"
	got = string(mime.Attachments[0].Content())
	if !strings.Contains(got, want) {
		t.Errorf("Content(): %q should contain: %q", got, want)
	}
}

func TestParseAttachmentOctet(t *testing.T) {
	msg := readMessage("attachment-octet.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}

	want := "A text section"
	if !strings.Contains(mime.Text, want) {
		t.Errorf("Text: %q should contain: %q", mime.Text, want)
	}
	if mime.HTML != "" {
		t.Error("mime.HTML should be empty, attachment is not for display, got:", mime.HTML)
	}
	if len(mime.Inlines) > 0 {
		t.Error("Should have no inlines, got:", len(mime.Inlines))
	}
	if len(mime.Attachments) != 1 {
		t.Fatal("Should have a single attachment, got:", len(mime.Attachments))
	}

	want = "ATTACHMENT.EXE"
	got := mime.Attachments[0].FileName()
	if got != want {
		t.Error("FileName() got:", got, "want:", want)
	}

	wantBytes := []byte{
		0x3, 0x17, 0xe1, 0x7e, 0xe8, 0xeb, 0xa2, 0x96, 0x9d, 0x95, 0xa7, 0x67, 0x82, 0x9,
		0xdf, 0x8e, 0xc, 0x2c, 0x6a, 0x2b, 0x9b, 0xbe, 0x79, 0xa4, 0x69, 0xd8, 0xae, 0x86,
		0xd7, 0xab, 0xa8, 0x72, 0x52, 0x15, 0xfb, 0x80, 0x8e, 0x47, 0xe1, 0xae, 0xaa, 0x5e,
		0xa2, 0xb2, 0xc0, 0x90, 0x59, 0xe3, 0x35, 0xf8, 0x60, 0xb7, 0xb1, 0x63, 0x77, 0xd7,
		0x5f, 0x92, 0x58, 0xa8, 0x75,
	}
	if !bytes.Equal(mime.Attachments[0].Content(), wantBytes) {
		t.Error("Attachment should have correct content")
	}
}

func TestParseAttachmentApplication(t *testing.T) {
	msg := readMessage("attachment-application.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if len(mime.Inlines) > 0 {
		t.Error("Should have no inlines, got:", len(mime.Inlines))
	}
	if len(mime.Attachments) != 1 {
		t.Fatal("Should have a single attachment, got:", len(mime.Attachments))
	}

	want := "some.doc"
	got := mime.Attachments[0].FileName()
	if got != want {
		t.Error("FileName() got:", got, "want:", want)
	}
}

func TestParseOtherParts(t *testing.T) {
	msg := readMessage("other-parts.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want := "A text section"
	if !strings.Contains(mime.Text, want) {
		t.Errorf("Text: %q should contain: %q", mime.Text, want)
	}
	if mime.HTML != "" {
		t.Error("mime.HTML should be empty, attachment is not for display, got:", mime.HTML)
	}
	if len(mime.Inlines) > 0 {
		t.Error("Should have no inlines, got:", len(mime.Inlines))
	}
	if len(mime.Attachments) > 0 {
		t.Fatal("Should have no attachments, got:", len(mime.Attachments))
	}
	if len(mime.OtherParts) != 1 {
		t.Fatal("Should have one other part, got:", len(mime.OtherParts))
	}

	want = "B05.gif"
	got := mime.OtherParts[0].FileName()
	if got != want {
		t.Error("FileName() got:", got, "want:", want)
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
	if !bytes.Equal(mime.OtherParts[0].Content(), wantBytes) {
		t.Error("Other part should have correct content")
	}
}

func TestParseInline(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be false")
	}

	want := "Test of text section"
	if !strings.Contains(mime.Text, want) {
		t.Errorf("Text: %q should contain: %q", mime.Text, want)
	}

	want = ">Test of HTML section<"
	if !strings.Contains(mime.HTML, want) {
		t.Errorf("HTML: %q should contain %q", mime.HTML, want)
	}

	if len(mime.Inlines) != 1 {
		t.Error("Should one inline, got:", len(mime.Inlines))
	}
	if len(mime.Attachments) > 0 {
		t.Fatal("Should have no attachments, got:", len(mime.Attachments))
	}

	want = "favicon.png"
	got := mime.Inlines[0].FileName()
	if got != want {
		t.Error("FileName() got:", got, "want:", want)
	}
	if !bytes.HasPrefix(mime.Inlines[0].Content(), []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("Inline should have correct content")
	}
}

func TestParseHTMLOnlyInline(t *testing.T) {
	msg := readMessage("html-only-inline.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	if !mime.IsTextFromHTML {
		t.Error("Expected text-from-HTML flag to be true")
	}

	want := "Test of HTML section"
	if !strings.Contains(mime.Text, want) {
		t.Errorf("Downconverted Text: %q should contain: %q", mime.Text, want)
	}

	want = ">Test of HTML section<"
	if !strings.Contains(mime.HTML, want) {
		t.Errorf("HTML: %q should contain %q", mime.HTML, want)
	}

	if len(mime.Inlines) != 1 {
		t.Error("Should one inline, got:", len(mime.Inlines))
	}
	if len(mime.Attachments) > 0 {
		t.Fatal("Should have no attachments, got:", len(mime.Attachments))
	}

	want = "favicon.png"
	got := mime.Inlines[0].FileName()
	if got != want {
		t.Error("FileName() got:", got, "want:", want)
	}
	if !bytes.HasPrefix(mime.Inlines[0].Content(), []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("Inline should have correct content")
	}
}

func TestParseNestedHeaders(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}
	if len(mime.Inlines) != 1 {
		t.Error("Should one inline, got:", len(mime.Inlines))
	}

	want := "favicon.png"
	got := mime.Inlines[0].FileName()
	if got != want {
		t.Error("FileName() got:", got, "want:", want)
	}
	want = "<8B8481A2-25CA-4886-9B5A-8EB9115DD064@skynet>"
	got = mime.Inlines[0].Header().Get("Content-Id")
	if got != want {
		t.Errorf("Content-Id header was: %q, want: %q", got, want)
	}
}

func TestParseEncodedSubjectAndAddress(t *testing.T) {
	// Even non-MIME messages should support encoded-words in headers
	// Also, encoded addresses should be suppored
	msg := readMessage("qp-ascii-header.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}

	want := "Test QP Subject!"
	got := mime.GetHeader("Subject")
	if got != want {
		t.Errorf("Subject was: %q, want: %q", got, want)
	}

	// Test UTF-8 subject line
	msg = readMessage("qp-utf8-header.raw")
	mime, err = ParseMIMEBody(msg)
	if err != nil {
		t.Fatal("Failed to parse MIME:", err)
	}

	want = "MIME UTF8 Test \u00a2 More Text"
	got = mime.GetHeader("Subject")
	if got != want {
		t.Errorf("Subject was: %q, want: %q", got, want)
	}

	toAddresses, err := mime.AddressList("To")
	if err != nil {
		t.Fatal("Failed to parse To list:", err)
	}
	if len(toAddresses) != 1 {
		t.Fatalf("len(toAddresses) == %v, want: %v", len(toAddresses), 1)
	}

	want = "Mirosław Marczak"
	got = toAddresses[0].Name
	if got != want {
		t.Errorf("To was: %q, want: %q", got, want)
	}
}

func TestDetectCharacterSetInHTML(t *testing.T) {
	msg := readMessage("non-mime-missing-charset.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatal("Failed to parse non-MIME:", err)
	}
	if strings.ContainsRune(mime.HTML, 0x80) {
		t.Error("HTML body should not have contained a Windows CP1250 Euro Symbol")
	}
	if !strings.ContainsRune(mime.HTML, 0x20ac) {
		t.Error("HTML body should have contained a Unicode Euro Symbol")
	}
}

func TestIsAttachment(t *testing.T) {
	var htests = []struct {
		want   bool
		header mail.Header
	}{
		{
			want: true,
			header: mail.Header{
				"Content-Disposition": []string{"attachment; filename=\"test.jpg\""}},
		},
		{
			want: true,
			header: mail.Header{
				"Content-Disposition": []string{"ATTACHMENT; filename=\"test.jpg\""}},
		},
		{
			want: true,
			header: mail.Header{
				"Content-Type":        []string{"image/jpg; name=\"test.jpg\""},
				"Content-Disposition": []string{"attachment; filename=\"test.jpg\""},
			},
		},
		{
			want: true,
			header: mail.Header{
				"Content-Type": []string{"attachment; filename=\"test.jpg\""}},
		},
		{
			want: true,
			header: mail.Header{
				"Content-Disposition": []string{"inline; filename=\"frog.jpg\""}},
		},
		{
			want: false,
			header: mail.Header{
				"Content-Disposition": []string{"non-attachment; filename=\"frog.jpg\""}},
		},
		{
			want:   false,
			header: mail.Header{},
		},
	}

	for _, s := range htests {
		got := IsAttachment(s.header)
		if got != s.want {
			t.Errorf("IsAttachment(%v) == %v, want: %v", s.header, got, s.want)
		}
	}
}

func TestIsPlain(t *testing.T) {
	var htests = []struct {
		want         bool
		header       mail.Header
		emptyIsPlain bool
	}{
		{
			want:         true,
			header:       mail.Header{"Content-Type": []string{"text/plain"}},
			emptyIsPlain: true,
		},
		{
			want:         true,
			header:       mail.Header{"Content-Type": []string{"text/html"}},
			emptyIsPlain: true,
		},
		{
			want:         true,
			header:       mail.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
			emptyIsPlain: true,
		},
		{
			want:         true,
			header:       mail.Header{},
			emptyIsPlain: true,
		},
		{
			want:         false,
			header:       mail.Header{},
			emptyIsPlain: false,
		},
		{
			want:         false,
			header:       mail.Header{"Content-Type": []string{"image/jpeg;"}},
			emptyIsPlain: true,
		},
		{
			want:         false,
			header:       mail.Header{"Content-Type": []string{"application/octet-stream"}},
			emptyIsPlain: true,
		},
	}

	for _, s := range htests {
		got := IsPlain(s.header, s.emptyIsPlain)
		if got != s.want {
			t.Errorf("IsPlain(%v, %v) == %v, want: %v", s.header, s.emptyIsPlain, got, s.want)
		}
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
		msg := readMessage(a.filename)
		m, err := ParseMIMEBody(msg)
		if err != nil {
			t.Fatal("Failed to parse MIME:", err)
		}
		if len(m.Attachments) != a.attachmentsLen {
			t.Fatal("len(Attachments) got:", len(m.Attachments), "want:", a.attachmentsLen)
		}
		if a.attachmentsLen > 0 &&
			!bytes.HasPrefix(m.Attachments[0].Content(), []byte{0x89, 'P', 'N', 'G'}) {
			t.Error("Content should be PNG image")
		}
		if len(m.Inlines) != a.inlinesLen {
			t.Fatal("len(Inlines) got:", len(m.Inlines), "want:", a.inlinesLen)
		}
		if a.inlinesLen > 0 &&
			!bytes.HasPrefix(m.Inlines[0].Content(), []byte{0x89, 'P', 'N', 'G'}) {
			t.Error("Content should be PNG image")
		}
	}
}

// readMessage is a test utility function to fetch a mail.Message object.
func readMessage(filename string) *mail.Message {
	// Open test email for parsing
	raw, err := os.Open(filepath.Join("testdata", "mail", filename))
	if err != nil {
		panic(fmt.Sprintf("Failed to open test data: %v", err))
	}

	// Parse email into a mail.Message object like we do
	reader := bufio.NewReader(raw)
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		panic(fmt.Sprintf("Failed to read message: %v", err))
	}

	return msg
}
