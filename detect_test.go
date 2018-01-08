package enmime

import (
	"net/textproto"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSinglePart(t *testing.T) {
	r, _ := os.Open(filepath.Join("testdata", "mail", "non-mime.raw"))
	msg, err := ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	if detectMultipartMessage(msg) {
		t.Error("Failed to identify non-multipart message")
	}
}

func TestDetectMultiPart(t *testing.T) {
	r, _ := os.Open(filepath.Join("testdata", "mail", "html-mime-inline.raw"))
	msg, err := ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	if !detectMultipartMessage(msg) {
		t.Error("Failed to identify multipart MIME message")
	}
}

func TestDetectUnknownMultiPart(t *testing.T) {
	r, _ := os.Open(filepath.Join("testdata", "mail", "unknown-part-type.raw"))
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
		r, _ := os.Open(filepath.Join("testdata", "mail", tt.filename))
		root, err := ReadParts(r)
		if err != nil {
			t.Fatal(err)
		}

		if !detectBinaryBody(root) {
			t.Errorf("Failed to identify binary body %s in %q", tt.disposition, tt.filename)
		}
	}
}

func TestDetectAttachmentHeader(t *testing.T) {
	var htests = []struct {
		want   bool
		header textproto.MIMEHeader
	}{
		{
			want: true,
			header: textproto.MIMEHeader{
				"Content-Disposition": []string{"attachment; filename=\"test.jpg\""}},
		},
		{
			want: true,
			header: textproto.MIMEHeader{
				"Content-Disposition": []string{"ATTACHMENT; filename=\"test.jpg\""}},
		},
		{
			want: true,
			header: textproto.MIMEHeader{
				"Content-Type":        []string{"image/jpg; name=\"test.jpg\""},
				"Content-Disposition": []string{"attachment; filename=\"test.jpg\""},
			},
		},
		{
			want: true,
			header: textproto.MIMEHeader{
				"Content-Type": []string{"attachment; filename=\"test.jpg\""}},
		},
		{
			want: true,
			header: textproto.MIMEHeader{
				"Content-Disposition": []string{"inline; filename=\"frog.jpg\""}},
		},
		{
			want: false,
			header: textproto.MIMEHeader{
				"Content-Disposition": []string{"non-attachment; filename=\"frog.jpg\""}},
		},
		{
			want:   false,
			header: textproto.MIMEHeader{},
		},
	}

	for _, s := range htests {
		got := detectAttachmentHeader(s.header)
		if got != s.want {
			t.Errorf("detectAttachmentHeader(%v) == %v, want: %v", s.header, got, s.want)
		}
	}
}

func TestDetectTextHeader(t *testing.T) {
	var htests = []struct {
		want         bool
		header       textproto.MIMEHeader
		emptyIsPlain bool
	}{
		{
			want:         true,
			header:       textproto.MIMEHeader{"Content-Type": []string{"text/plain"}},
			emptyIsPlain: true,
		},
		{
			want:         true,
			header:       textproto.MIMEHeader{"Content-Type": []string{"text/html"}},
			emptyIsPlain: true,
		},
		{
			want:         true,
			header:       textproto.MIMEHeader{"Content-Type": []string{"text/html; charset=utf-8"}},
			emptyIsPlain: true,
		},
		{
			want:         true,
			header:       textproto.MIMEHeader{},
			emptyIsPlain: true,
		},
		{
			want:         false,
			header:       textproto.MIMEHeader{},
			emptyIsPlain: false,
		},
		{
			want:         false,
			header:       textproto.MIMEHeader{"Content-Type": []string{"image/jpeg;"}},
			emptyIsPlain: true,
		},
		{
			want:         false,
			header:       textproto.MIMEHeader{"Content-Type": []string{"application/octet-stream"}},
			emptyIsPlain: true,
		},
	}

	for _, s := range htests {
		got := detectTextHeader(s.header, s.emptyIsPlain)
		if got != s.want {
			t.Errorf("detectTextHeader(%v, %v) == %v, want: %v",
				s.header, s.emptyIsPlain, got, s.want)
		}
	}
}
