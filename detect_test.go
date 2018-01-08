package enmime

import "testing"

func TestDetectSinglePart(t *testing.T) {
	r := openTestData("mail", "non-mime.raw")
	msg, err := ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	if detectMultipartMessage(msg) {
		t.Error("Failed to identify non-multipart message")
	}
}

func TestDetectMultiPart(t *testing.T) {
	r := openTestData("mail", "html-mime-inline.raw")
	msg, err := ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	if !detectMultipartMessage(msg) {
		t.Error("Failed to identify multipart MIME message")
	}
}

func TestDetectUnknownMultiPart(t *testing.T) {
	r := openTestData("mail", "unknown-part-type.raw")
	msg, err := ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	if !detectMultipartMessage(msg) {
		t.Error("Failed to identify multipart MIME message of unknown type")
	}
}

func TestDetectBinaryBody(t *testing.T) {
	ttable := []struct {
		filename    string
		disposition string
	}{
		{filename: "attachment-only.raw", disposition: "attachment"},
		{filename: "attachment-only-inline.raw", disposition: "inline"},
		{filename: "attachment-only-no-disposition.raw", disposition: "none"},
	}
	for _, tt := range ttable {
		r := openTestData("mail", tt.filename)
		root, err := ReadParts(r)
		if err != nil {
			t.Fatal(err)
		}

		if !detectBinaryBody(root) {
			t.Errorf("Failed to identify binary body %s in %q", tt.disposition, tt.filename)
		}
	}
}
