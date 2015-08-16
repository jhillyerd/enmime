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

	assert.Equal(t, "Test of text section", mime.Text)
}

func TestParseMultiMixedText(t *testing.T) {
	msg := readMessage("mime-mixed.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Equal(t, "Section one\n\n--\nSection two", mime.Text,
		"Text parts should be concatenated")
}

func TestParseMultiSignedText(t *testing.T) {
	msg := readMessage("mime-signed.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Equal(t, "Section one\n\n--\nSection two", mime.Text,
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
	assert.Equal(t, "", mime.Html, "Html attachment is not for display")
	assert.Equal(t, 0, len(mime.Inlines), "Should have no inlines")
	assert.Equal(t, 1, len(mime.Attachments), "Should have a single attachment")
	assert.Equal(t, "test.html", mime.Attachments[0].FileName(), "Attachment should have correct filename")
	assert.Contains(t, string(mime.Attachments[0].Content()), "<html>",
		"Attachment should have correct content")

	//for _, a := range mime.Attachments {
	//	fmt.Printf("%v %v\n", a.ContentType(), a.Disposition())
	//}
}

func TestParseAttachmentOctet(t *testing.T) {
	msg := readMessage("attachment-octet.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Contains(t, mime.Text, "A text section")
	assert.Equal(t, "", mime.Html, "Html attachment is not for display")
	assert.Equal(t, 0, len(mime.Inlines), "Should have no inlines")
	assert.Equal(t, 1, len(mime.Attachments), "Should have a single attachment")
	assert.Equal(t, "ATTACHMENT.EXE", mime.Attachments[0].FileName(), "Attachment should have correct filename")
	assert.Equal(t,
		[]byte{0x3, 0x17, 0xe1, 0x7e, 0xe8, 0xeb, 0xa2, 0x96, 0x9d, 0x95, 0xa7, 0x67, 0x82, 0x9, 0xdf, 0x8e, 0xc, 0x2c, 0x6a, 0x2b, 0x9b, 0xbe, 0x79, 0xa4, 0x69, 0xd8, 0xae, 0x86, 0xd7, 0xab, 0xa8, 0x72, 0x52, 0x15, 0xfb, 0x80, 0x8e, 0x47, 0xe1, 0xae, 0xaa, 0x5e, 0xa2, 0xb2, 0xc0, 0x90, 0x59, 0xe3, 0x35, 0xf8, 0x60, 0xb7, 0xb1, 0x63, 0x77, 0xd7, 0x5f, 0x92, 0x58, 0xa8, 0x75}, mime.Attachments[0].Content(),
		"Attachment should have correct content")

}

func TestParseOtherParts(t *testing.T) {
	msg := readMessage("other-parts.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Contains(t, mime.Text, "A text section")
	assert.Equal(t, "", mime.Html, "No Html attachment available")
	assert.Equal(t, 0, len(mime.Inlines), "Should have no inlines")
	assert.Equal(t, 0, len(mime.Attachments), "Should have no attachment")
	assert.Equal(t, 1, len(mime.OtherParts), "Should have one OtherParts")
	assert.Equal(t, "B05.gif", mime.OtherParts[0].FileName(), "Part should have correct filename")
	assert.Equal(t,
		[]byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0xf, 0x0, 0xf, 0x0, 0xa2, 0x5, 0x0, 0xde, 0xeb, 0xf3, 0x5b, 0xb0, 0xec, 0x0, 0x89, 0xe3, 0xa3, 0xd0, 0xed, 0x0, 0x46, 0x74, 0xdd, 0xed, 0xfa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x21, 0xf9, 0x4, 0x1, 0x0, 0x0, 0x5, 0x0, 0x2c, 0x0, 0x0, 0x0, 0x0, 0xf, 0x0, 0xf, 0x0, 0x0, 0x3, 0x40, 0x58, 0x25, 0xa4, 0x4b, 0xb0, 0x39, 0x1, 0x46, 0xa3, 0x23, 0x5b, 0x47, 0x46, 0x68, 0x9d, 0x20, 0x6, 0x9f, 0xd2, 0x95, 0x45, 0x44, 0x8, 0xe8, 0x29, 0x39, 0x69, 0xeb, 0xbd, 0xc, 0x41, 0x4a, 0xae, 0x82, 0xcd, 0x1c, 0x9f, 0xce, 0xaf, 0x1f, 0xc3, 0x34, 0x18, 0xc2, 0x42, 0xb8, 0x80, 0xf1, 0x18, 0x84, 0xc0, 0x9e, 0xd0, 0xe8, 0xf2, 0x1, 0xb5, 0x19, 0xad, 0x41, 0x53, 0x33, 0x9b, 0x0, 0x0, 0x3b}, mime.OtherParts[0].Content(),
		"Part should have correct content")
}

func TestParseInline(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Contains(t, mime.Text, "Test of text section", "Should have text section")
	assert.Contains(t, mime.Html, ">Test of HTML section<", "Should have html section")
	assert.Equal(t, 1, len(mime.Inlines), "Should have one inline")
	assert.Equal(t, 0, len(mime.Attachments), "Should have no attachments")
	assert.Equal(t, "favicon.png", mime.Inlines[0].FileName(), "Inline should have correct filename")
	assert.True(t, bytes.HasPrefix(mime.Inlines[0].Content(), []byte{0x89, 'P', 'N', 'G'}),
		"Content should be PNG image")
}

func TestParseNestedHeaders(t *testing.T) {
	msg := readMessage("html-mime-inline.raw")
	mime, err := ParseMIMEBody(msg)

	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}

	assert.Equal(t, 1, len(mime.Inlines), "Should have one inline")
	assert.Equal(t, "favicon.png", mime.Inlines[0].FileName(), "Inline should have correct filename")
	assert.Equal(t, "<8B8481A2-25CA-4886-9B5A-8EB9115DD064@skynet>", mime.Inlines[0].Header().Get("Content-Id"), "Inline should have a Content-Id header")
}

func TestParseEncodedSubjectAndAddress(t *testing.T) {
	// Even non-MIME messages should support encoded-words in headers
	// Also, encoded addresses should be suppored
	msg := readMessage("qp-ascii-header.raw")
	mime, err := ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse non-MIME: %v", err)
	}
	assert.Equal(t, "Test QP Subject!", mime.GetHeader("Subject"))

	// Test UTF-8 subject line
	msg = readMessage("qp-utf8-header.raw")
	mime, err = ParseMIMEBody(msg)
	if err != nil {
		t.Fatalf("Failed to parse MIME: %v", err)
	}
	assert.Equal(t, "MIME UTF8 Test \u00a2 More Text", mime.GetHeader("Subject"))
	toAddresses, err := mime.AddressList("To")
	if err != nil {
		t.Fatalf("Failed to parse To list: %v", err)
	}
	assert.Equal(t, 1, len(toAddresses))
	assert.Equal(t, "Mirosław Marczak", toAddresses[0].Name)
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
