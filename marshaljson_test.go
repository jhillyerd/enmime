package enmime_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime/v2"
)

const jsonTestMessage = "From: sender@example.com\r\n" +
	"To: recipient@example.com\r\n" +
	"Subject: json test\r\n" +
	"Content-Type: multipart/mixed; boundary=BOUNDARY\r\n" +
	"\r\n" +
	"--BOUNDARY\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n" +
	"plain body\r\n" +
	"--BOUNDARY\r\n" +
	"Content-Type: text/html\r\n" +
	"\r\n" +
	"<p>html body</p>\r\n" +
	"--BOUNDARY\r\n" +
	"Content-Type: application/octet-stream\r\n" +
	"Content-Disposition: attachment; filename=\"data.bin\"\r\n" +
	"\r\n" +
	"attachment bytes\r\n" +
	"--BOUNDARY--\r\n"

func TestEnvelopeMarshalJSON(t *testing.T) {
	env, err := enmime.ReadEnvelope(strings.NewReader(jsonTestMessage))
	if err != nil {
		t.Fatal("Failed to parse message:", err)
	}

	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal("Failed to marshal envelope:", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal("Failed to unmarshal envelope JSON:", err)
	}

	// Root is intentionally left out so the whole part tree isn't duplicated
	// alongside the flat attachment/inline/other part lists.
	if _, ok := got["Root"]; ok {
		t.Error("Expected Root to be omitted from envelope JSON")
	}

	var decoded struct {
		Header      map[string][]string `json:"header"`
		Text        string              `json:"text"`
		HTML        string              `json:"html"`
		Attachments []struct {
			FileName    string `json:"fileName"`
			ContentType string `json:"contentType"`
		} `json:"attachments"`
	}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal("Failed to decode envelope JSON:", err)
	}

	if decoded.Text != "plain body" {
		t.Errorf("Text: got %q, want %q", decoded.Text, "plain body")
	}
	if !strings.Contains(decoded.HTML, "html body") {
		t.Errorf("HTML: got %q, want it to contain %q", decoded.HTML, "html body")
	}
	if got := decoded.Header["Subject"]; len(got) != 1 || got[0] != "json test" {
		t.Errorf("Header Subject: got %v, want [json test]", got)
	}
	if len(decoded.Attachments) != 1 {
		t.Fatalf("Attachments: got %d, want 1", len(decoded.Attachments))
	}
	if decoded.Attachments[0].FileName != "data.bin" {
		t.Errorf("Attachment FileName: got %q, want %q", decoded.Attachments[0].FileName, "data.bin")
	}
	if decoded.Attachments[0].ContentType != "application/octet-stream" {
		t.Errorf("Attachment ContentType: got %q, want %q",
			decoded.Attachments[0].ContentType, "application/octet-stream")
	}
}

// TestPartMarshalJSON confirms a Part marshals without tripping over the Parent
// back-pointer that otherwise produces a json cycle error.
func TestPartMarshalJSON(t *testing.T) {
	env, err := enmime.ReadEnvelope(strings.NewReader(jsonTestMessage))
	if err != nil {
		t.Fatal("Failed to parse message:", err)
	}
	if len(env.Attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(env.Attachments))
	}

	b, err := json.Marshal(env.Attachments[0])
	if err != nil {
		t.Fatal("Failed to marshal part:", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal("Failed to unmarshal part JSON:", err)
	}
	for _, link := range []string{"parent", "firstChild", "nextSibling", "Parent"} {
		if _, ok := got[link]; ok {
			t.Errorf("Expected %s to be omitted from part JSON", link)
		}
	}
	if _, ok := got["fileName"]; !ok {
		t.Error("Expected fileName in part JSON")
	}
}
