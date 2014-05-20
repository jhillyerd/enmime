package enmime

import (
	"bufio"
	"bytes"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifySinglePart(t *testing.T) {
	msg := readMessage("non-mime.raw")
	assert.False(t, IsMultipartMessage(msg), "Failed to identify non-multipart message")
}

func TestIdentifyMultiPart(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	assert.True(t, IsMultipartMessage(msg), "Failed to identify multipart MIME message")
}

func TestParseNonMime(t *testing.T) {
	msg := readMessage("non-mime.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse non-MIME: %v", err)
	}

	assert.Contains(t, mime.Text, "This is a test mailing")
	assert.Empty(t, mime.Html, "Expected no HTML body")
}

func TestParseNonMimeHtml(t *testing.T) {
	msg := readMessage("non-mime-html.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse non-MIME: %v", err)
	}

	assert.Contains(t, mime.Text, "This is a test mailing")
	assert.Contains(t, mime.Html, "This is a test mailing")
}

func TestParseMimeTree(t *testing.T) {
	msg := readMessage("attachment.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.NotNil(t, mime.Root, "Message should have a root node")
}

func TestParseInlineText(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Equal(t, mime.Text, "Test of text section")
}

func TestParseMultiMixedText(t *testing.T) {
	msg := readMessage("mime-mixed.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Equal(t, mime.Text, "Section one\n\n--\nSection two",
		"Text parts should be concatenated")
}

func TestParseMultiSignedText(t *testing.T) {
	msg := readMessage("mime-signed.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Equal(t, mime.Text, "Section one\n\n--\nSection two",
		"Text parts should be concatenated")
}

func TestParseQuotedPrintable(t *testing.T) {
	msg := readMessage("quoted-printable.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Contains(t, mime.Text, "Phasellus sit amet arcu")
}

func TestParseQuotedPrintableMime(t *testing.T) {
	msg := readMessage("quoted-printable-mime.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Contains(t, mime.Text, "Nullam venenatis ante")
}

func TestParseInlineHtml(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Contains(t, mime.Html, "<html>")
	assert.Contains(t, mime.Html, "Test of HTML section")
}

func TestParseAttachment(t *testing.T) {
	msg := readMessage("attachment.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Contains(t, mime.Text, "A text section")
	assert.Equal(t, mime.Html, "", "Html attachment is not for display")
	assert.Equal(t, len(mime.Inlines), 0, "Should have no inlines")
	assert.Equal(t, len(mime.Attachments), 1, "Should have a single attachment")
	assert.Equal(t, mime.Attachments[0].FileName(), "test.html", "Attachment should have correct filename")
	assert.Contains(t, string(mime.Attachments[0].Content()), "<html>",
		"Attachment should have correct content")

	//for _, a := range mime.Attachments {
	//	fmt.Printf("%v %v\n", a.ContentType(), a.Disposition())
	//}
}

func TestParseInline(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Contains(t, mime.Text, "Test of text section", "Should have text section")
	assert.Contains(t, mime.Html, ">Test of HTML section<", "Should have html section")
	assert.Equal(t, len(mime.Inlines), 1, "Should have one inline")
	assert.Equal(t, len(mime.Attachments), 0, "Should have no attachments")
	assert.Equal(t, mime.Inlines[0].FileName(), "favicon.png", "Inline should have correct filename")
	assert.True(t, bytes.HasPrefix(mime.Inlines[0].Content(), []byte{0x89, 'P', 'N', 'G'}),
		"Content should be PNG image")
}

func TestParseNestedHeaders(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Equal(t, len(mime.Inlines), 1, "Should have one inline")
	assert.Equal(t, mime.Inlines[0].FileName(), "favicon.png", "Inline should have correct filename")
	assert.Equal(t, mime.Inlines[0].Header().Get("Content-Id"), "<8B8481A2-25CA-4886-9B5A-8EB9115DD064@skynet>", "Inline should have a Content-Id header")
}

func TestParseEncodedSubject(t *testing.T) {
	// Even non-MIME messages should support encoded-words in headers
	msg := readMessage("qp-ascii-header.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse non-MIME: %v", err)
	}
	assert.Equal(t, mime.GetHeader("Subject"), "Test QP Subject!")

	// Test UTF-8 subject line
	msg = readMessage("qp-utf8-header.raw")
	mime, err = ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}
	assert.Equal(t, mime.GetHeader("Subject"), "MIME UTF8 Test \u00a2 More Text")
}

// readMessage is a test utility function to fetch a mail.Message object.
func readMessage(filename string) *mail.Message {
	// Open test email for parsing
	raw, err := os.Open(filepath.Join("test-data", "mail", filename))
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
