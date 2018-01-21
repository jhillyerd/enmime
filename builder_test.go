package enmime_test

import (
	"reflect"
	"testing"

	"github.com/jhillyerd/enmime"
)

var strSlice = []string{"word"}

func TestBuilderEquals(t *testing.T) {
	var a, b *enmime.MailBuilder

	if !a.Equals(b) {
		t.Error("nil PartBuilders should be equal")
	}

	a = enmime.Builder()
	b = enmime.Builder()
	if !a.Equals(b) {
		t.Error("New PartBuilders should be equal")
	}
}

func TestBuilderFrom(t *testing.T) {
	a := enmime.Builder().From("same")
	b := enmime.Builder().From("same")
	if !a.Equals(b) {
		t.Error("Same From(value) should be equal")
	}

	a = enmime.Builder().From("foo")
	b = enmime.Builder().From("bar")
	if a.Equals(b) {
		t.Error("Different From(value) should not be equal")
	}

	a = enmime.Builder().From("foo")
	b = a.From("bar")
	if a.Equals(b) {
		t.Error("From() should not mutate receiver, failed")
	}

	want := "from@inbucket.org"
	a = enmime.Builder().From(want).Subject("foo").To(strSlice)
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header.Get("From")
	if got != want {
		t.Errorf("From: %q, want: %q", got, want)
	}
}

func TestBuilderSubject(t *testing.T) {
	a := enmime.Builder().Subject("same")
	b := enmime.Builder().Subject("same")
	if !a.Equals(b) {
		t.Error("Same Subject(value) should be equal")
	}

	a = enmime.Builder().Subject("foo")
	b = enmime.Builder().Subject("bar")
	if a.Equals(b) {
		t.Error("Different Subject(value) should not be equal")
	}

	a = enmime.Builder().Subject("foo")
	b = a.Subject("bar")
	if a.Equals(b) {
		t.Error("Subject() should not mutate receiver, failed")
	}

	want := "engaging subject"
	a = enmime.Builder().Subject(want).From("foo").To(strSlice)
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header.Get("Subject")
	if got != want {
		t.Errorf("Subject: %q, want: %q", got, want)
	}
}

func TestBuilderTo(t *testing.T) {
	a := enmime.Builder().To([]string{"same"})
	b := enmime.Builder().To([]string{"same"})
	if !a.Equals(b) {
		t.Error("Same To(value) should be equal")
	}

	a = enmime.Builder().To([]string{"foo"})
	b = enmime.Builder().To([]string{"bar"})
	if a.Equals(b) {
		t.Error("Different To(value) should not be equal")
	}

	a = enmime.Builder().To([]string{"foo"})
	b = a.To([]string{"bar"})
	if a.Equals(b) {
		t.Error("To() should not mutate receiver, failed")
	}

	input := []string{"one@inbucket.org", "two@inbucket.org"}
	want := "one@inbucket.org, two@inbucket.org"
	a = enmime.Builder().To(input).From("foo").Subject("foo")
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header.Get("To")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("To: %q, want: %q", got, want)
	}
}

func TestBuilderCC(t *testing.T) {
	a := enmime.Builder().CC([]string{"same"})
	b := enmime.Builder().CC([]string{"same"})
	if !a.Equals(b) {
		t.Error("Same CC(value) should be equal")
	}

	a = enmime.Builder().CC([]string{"foo"})
	b = enmime.Builder().CC([]string{"bar"})
	if a.Equals(b) {
		t.Error("Different CC(value) should not be equal")
	}

	a = enmime.Builder().CC([]string{"foo"})
	b = a.CC([]string{"bar"})
	if a.Equals(b) {
		t.Error("CC() should not mutate receiver, failed")
	}

	input := []string{"one@inbucket.org", "two@inbucket.org"}
	want := "one@inbucket.org, two@inbucket.org"
	a = enmime.Builder().CC(input).From("foo").Subject("foo")
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header.Get("Cc")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CC: %q, want: %q", got, want)
	}
}

func TestBuilderBCC(t *testing.T) {
	a := enmime.Builder().BCC([]string{"same"})
	b := enmime.Builder().BCC([]string{"same"})
	if !a.Equals(b) {
		t.Error("Same BCC(value) should be equal")
	}

	a = enmime.Builder().BCC([]string{"foo"})
	b = enmime.Builder().BCC([]string{"bar"})
	if a.Equals(b) {
		t.Error("Different BCC(value) should not be equal")
	}

	a = enmime.Builder().BCC([]string{"foo"})
	b = a.BCC([]string{"bar"})
	if a.Equals(b) {
		t.Error("BCC() should not mutate receiver, failed")
	}

	// BCC doesn't show up in headers
	input := []string{"one@inbucket.org", "two@inbucket.org"}
	want := ""
	a = enmime.Builder().BCC(input).From("foo").Subject("foo")
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header.Get("Bcc")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BCC: %q, want: %q", got, want)
	}
}

func TestBuilderText(t *testing.T) {
	a := enmime.Builder().Text([]byte("same"))
	b := enmime.Builder().Text([]byte("same"))
	if !a.Equals(b) {
		t.Error("Same Text(value) should be equal")
	}

	a = enmime.Builder().Text([]byte("foo"))
	b = enmime.Builder().Text([]byte("bar"))
	if a.Equals(b) {
		t.Error("Different Text(value) should not be equal")
	}

	a = enmime.Builder().Text([]byte("foo"))
	b = a.Text([]byte("bar"))
	if a.Equals(b) {
		t.Error("Text() should not mutate receiver, failed")
	}

	want := "test text body"
	a = enmime.Builder().Text([]byte(want)).From("foo").Subject("foo").To(strSlice)
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := string(p.Content)
	if got != want {
		t.Errorf("Content: %q, want: %q", got, want)
	}
	want = "text/plain"
	got = p.ContentType
	if got != want {
		t.Errorf("Content-Type: %q, want: %q", got, want)
	}
	want = "utf-8"
	got = p.Charset
	if got != want {
		t.Errorf("Charset: %q, want: %q", got, want)
	}
}

func TestBuilderHTML(t *testing.T) {
	a := enmime.Builder().HTML([]byte("same"))
	b := enmime.Builder().HTML([]byte("same"))
	if !a.Equals(b) {
		t.Error("Same HTML(value) should be equal")
	}

	a = enmime.Builder().HTML([]byte("foo"))
	b = enmime.Builder().HTML([]byte("bar"))
	if a.Equals(b) {
		t.Error("Different HTML(value) should not be equal")
	}

	a = enmime.Builder().HTML([]byte("foo"))
	b = a.HTML([]byte("bar"))
	if a.Equals(b) {
		t.Error("HTML() should not mutate receiver, failed")
	}

	want := "test html body"
	a = enmime.Builder().HTML([]byte(want)).From("foo").Subject("foo").To(strSlice)
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := string(p.Content)
	if got != want {
		t.Errorf("Content: %q, want: %q", got, want)
	}
	want = "text/html"
	got = p.ContentType
	if got != want {
		t.Errorf("Content-Type: %q, want: %q", got, want)
	}
	want = "utf-8"
	got = p.Charset
	if got != want {
		t.Errorf("Charset: %q, want: %q", got, want)
	}
}

func TestBuilderMultiBody(t *testing.T) {
	text := "test text body"
	html := "test html body"
	a := enmime.Builder().
		Text([]byte(text)).
		HTML([]byte(html)).
		From("foo").
		Subject("foo").
		To(strSlice)
	root, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}

	// Should be multipart
	p := root
	want := "multipart/alternative"
	got := p.ContentType
	if got != want {
		t.Errorf("Content-Type: %q, want: %q", got, want)
	}

	// Find text part
	p = root.DepthMatchFirst(func(p *enmime.Part) bool { return p.ContentType == "text/plain" })
	if p == nil {
		t.Fatal("Did not find a text/plain part")
	}
	want = text
	got = string(p.Content)
	if got != want {
		t.Errorf("Content: %q, want: %q", got, want)
	}
	want = "utf-8"
	got = p.Charset
	if got != want {
		t.Errorf("Charset: %q, want: %q", got, want)
	}

	// Find HTML part
	p = root.DepthMatchFirst(func(p *enmime.Part) bool { return p.ContentType == "text/html" })
	if p == nil {
		t.Fatal("Did not find a text/html part")
	}
	want = html
	got = string(p.Content)
	if got != want {
		t.Errorf("Content: %q, want: %q", got, want)
	}
	want = "utf-8"
	got = p.Charset
	if got != want {
		t.Errorf("Charset: %q, want: %q", got, want)
	}
}
