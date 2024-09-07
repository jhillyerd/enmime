package enmime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jhillyerd/enmime/v2/internal/textproto"
)

func TestDetectSinglePart(t *testing.T) {
	r, _ := os.Open(filepath.Join("testdata", "mail", "non-mime.raw"))
	msg, err := ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	if detectMultipartMessage(msg, false) {
		t.Error("Failed to identify non-multipart message")
	}
}

func TestDetectMultiPart(t *testing.T) {
	r, _ := os.Open(filepath.Join("testdata", "mail", "html-mime-inline.raw"))
	msg, err := ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	if !detectMultipartMessage(msg, false) {
		t.Error("Failed to identify multipart MIME message")
	}
}

func TestDetectUnknownMultiPart(t *testing.T) {
	r, _ := os.Open(filepath.Join("testdata", "mail", "unknown-part-type.raw"))
	msg, err := ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	if !detectMultipartMessage(msg, false) {
		t.Error("Failed to identify multipart MIME message of unknown type")
	}
}

func TestDetectMultipartWithoutBoundary(t *testing.T) {
	r, _ := os.Open(filepath.Join("testdata", "mail", "multipart-wo-boundary.raw"))
	msg, err := ReadParts(r)
	if err != nil {
		t.Fatal(err)
	}

	if !detectMultipartMessage(msg, false) {
		t.Error("Failed to identify multipart MIME message")
	}

	if detectMultipartMessage(msg, true) {
		t.Error("Failed to identify multipart MIME message without boundaries as single-part")
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
		{filename: "attachment-only-text-attachment.raw", disposition: "attachment"},
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
			want: false,
			header: textproto.MIMEHeader{
				"Content-Disposition": []string{"inline"}},
		},
		{
			want: false,
			header: textproto.MIMEHeader{
				"Content-Disposition": []string{"inline; broken"}},
		},
		{
			want: true,
			header: textproto.MIMEHeader{
				"Content-Disposition": []string{"attachment; broken"}},
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

	root := &Part{parser: &defaultParser}

	for _, s := range htests {
		got := detectAttachmentHeader(root, s.header)
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

	root := &Part{parser: &defaultParser}

	for _, s := range htests {
		got := detectTextHeader(root, s.header, s.emptyIsPlain)
		if got != s.want {
			t.Errorf("detectTextHeader(%v, %v) == %v, want: %v",
				s.header, s.emptyIsPlain, got, s.want)
		}
	}
}
