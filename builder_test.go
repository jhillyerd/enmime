package enmime_test

import (
	"net/mail"
	"reflect"
	"testing"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/internal/test"
)

var addrSlice = []mail.Address{{Name: "name", Address: "addr"}}

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
	a := enmime.Builder().From("name", "same")
	b := enmime.Builder().From("name", "same")
	if !a.Equals(b) {
		t.Error("Same From(value) should be equal")
	}

	a = enmime.Builder().From("name", "foo")
	b = enmime.Builder().From("name", "bar")
	if a.Equals(b) {
		t.Error("Different From(value) should not be equal")
	}

	a = enmime.Builder().From("name", "foo")
	b = a.From("name", "bar")
	if a.Equals(b) {
		t.Error("From() should not mutate receiver, failed")
	}

	want := mail.Address{Name: "name", Address: "from@inbucket.org"}
	a = enmime.Builder().From(want.Name, want.Address).Subject("foo").ToAddrs(addrSlice)
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header.Get("From")
	if got != want.String() {
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
	a = enmime.Builder().Subject(want).From("name", "foo").ToAddrs(addrSlice)
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
	a := enmime.Builder().To("name", "same")
	b := enmime.Builder().To("name", "same")
	if !a.Equals(b) {
		t.Error("Same To(value) should be equal")
	}

	a = enmime.Builder().To("name", "foo")
	b = enmime.Builder().To("name", "bar")
	if a.Equals(b) {
		t.Error("Different To(value) should not be equal")
	}

	a = enmime.Builder().To("name", "foo")
	b = a.To("name", "bar")
	if a.Equals(b) {
		t.Error("To() should not mutate receiver, failed")
	}

	a = enmime.Builder().From("name", "foo").Subject("foo")
	a = a.To("one", "one@inbucket.org")
	a = a.To("two", "two@inbucket.org")
	want := "\"one\" <one@inbucket.org>, \"two\" <two@inbucket.org>"
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header.Get("To")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("To: %q, want: %q", got, want)
	}
}

func TestBuilderToAddrs(t *testing.T) {
	a := enmime.Builder().ToAddrs([]mail.Address{{Name: "name", Address: "same"}})
	b := enmime.Builder().ToAddrs([]mail.Address{{Name: "name", Address: "same"}})
	if !a.Equals(b) {
		t.Error("Same To(value) should be equal")
	}

	a = enmime.Builder().ToAddrs([]mail.Address{{Name: "name", Address: "foo"}})
	b = enmime.Builder().ToAddrs([]mail.Address{{Name: "name", Address: "bar"}})
	if a.Equals(b) {
		t.Error("Different To(value) should not be equal")
	}

	a = enmime.Builder().ToAddrs([]mail.Address{{Name: "name", Address: "foo"}})
	b = a.ToAddrs([]mail.Address{{Name: "name", Address: "bar"}})
	if a.Equals(b) {
		t.Error("To() should not mutate receiver, failed")
	}

	input := []mail.Address{
		{Name: "one", Address: "one@inbucket.org"},
		{Name: "two", Address: "two@inbucket.org"},
	}
	want := "\"one\" <one@inbucket.org>, \"two\" <two@inbucket.org>"
	a = enmime.Builder().ToAddrs(input).From("name", "foo").Subject("foo")
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
	a := enmime.Builder().CC("name", "same")
	b := enmime.Builder().CC("name", "same")
	if !a.Equals(b) {
		t.Error("Same CC(value) should be equal")
	}

	a = enmime.Builder().CC("name", "foo")
	b = enmime.Builder().CC("name", "bar")
	if a.Equals(b) {
		t.Error("Different CC(value) should not be equal")
	}

	a = enmime.Builder().CC("name", "foo")
	b = a.CC("name", "bar")
	if a.Equals(b) {
		t.Error("CC() should not mutate receiver, failed")
	}

	a = enmime.Builder().From("name", "foo").Subject("foo")
	a = a.CC("one", "one@inbucket.org")
	a = a.CC("two", "two@inbucket.org")
	want := "\"one\" <one@inbucket.org>, \"two\" <two@inbucket.org>"
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header.Get("CC")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CC: %q, want: %q", got, want)
	}
}

func TestBuilderCCAddrs(t *testing.T) {
	a := enmime.Builder().CCAddrs([]mail.Address{{Name: "name", Address: "same"}})
	b := enmime.Builder().CCAddrs([]mail.Address{{Name: "name", Address: "same"}})
	if !a.Equals(b) {
		t.Error("Same CC(value) should be equal")
	}

	a = enmime.Builder().CCAddrs([]mail.Address{{Name: "name", Address: "foo"}})
	b = enmime.Builder().CCAddrs([]mail.Address{{Name: "name", Address: "bar"}})
	if a.Equals(b) {
		t.Error("Different CC(value) should not be equal")
	}

	a = enmime.Builder().CCAddrs([]mail.Address{{Name: "name", Address: "foo"}})
	b = a.CCAddrs([]mail.Address{{Name: "name", Address: "bar"}})
	if a.Equals(b) {
		t.Error("CC() should not mutate receiver, failed")
	}

	input := []mail.Address{
		{Name: "one", Address: "one@inbucket.org"},
		{Name: "two", Address: "two@inbucket.org"},
	}
	want := "\"one\" <one@inbucket.org>, \"two\" <two@inbucket.org>"
	a = enmime.Builder().CCAddrs(input).From("name", "foo").Subject("foo")
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
	a := enmime.Builder().BCC("name", "same")
	b := enmime.Builder().BCC("name", "same")
	if !a.Equals(b) {
		t.Error("Same BCC(value) should be equal")
	}

	a = enmime.Builder().BCC("name", "foo")
	b = enmime.Builder().BCC("name", "bar")
	if a.Equals(b) {
		t.Error("Different BCC(value) should not be equal")
	}

	a = enmime.Builder().BCC("name", "foo")
	b = a.BCC("name", "bar")
	if a.Equals(b) {
		t.Error("BCC() should not mutate receiver, failed")
	}

	a = enmime.Builder().From("name", "foo").Subject("foo")
	a = a.BCC("one", "one@inbucket.org")
	a = a.BCC("two", "two@inbucket.org")
	want := ""
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header.Get("BCC")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BCC: %q, want: %q", got, want)
	}
}

func TestBuilderBCCAddrs(t *testing.T) {
	a := enmime.Builder().BCCAddrs([]mail.Address{{Name: "name", Address: "same"}})
	b := enmime.Builder().BCCAddrs([]mail.Address{{Name: "name", Address: "same"}})
	if !a.Equals(b) {
		t.Error("Same BCC(value) should be equal")
	}

	a = enmime.Builder().BCCAddrs([]mail.Address{{Name: "name", Address: "foo"}})
	b = enmime.Builder().BCCAddrs([]mail.Address{{Name: "name", Address: "bar"}})
	if a.Equals(b) {
		t.Error("Different BCC(value) should not be equal")
	}

	a = enmime.Builder().BCCAddrs([]mail.Address{{Name: "name", Address: "foo"}})
	b = a.BCCAddrs([]mail.Address{{Name: "name", Address: "bar"}})
	if a.Equals(b) {
		t.Error("BCC() should not mutate receiver, failed")
	}

	// BCC doesn't show up in headers
	input := []mail.Address{
		{Name: "one", Address: "one@inbucket.org"},
		{Name: "two", Address: "two@inbucket.org"},
	}
	want := ""
	a = enmime.Builder().BCCAddrs(input).From("name", "foo").Subject("foo")
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
	a = enmime.Builder().Text([]byte(want)).From("name", "foo").Subject("foo").ToAddrs(addrSlice)
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
	a = enmime.Builder().HTML([]byte(want)).From("name", "foo").Subject("foo").ToAddrs(addrSlice)
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
		From("name", "foo").
		Subject("foo").
		ToAddrs(addrSlice)
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

func TestBuilderAddAttachment(t *testing.T) {
	a := enmime.Builder().AddAttachment([]byte("same"), "ct", "fn")
	b := enmime.Builder().AddAttachment([]byte("same"), "ct", "fn")
	if !a.Equals(b) {
		t.Error("Same AddAttachment(value) should be equal")
	}

	a = enmime.Builder().AddAttachment([]byte("foo"), "ct", "fn")
	b = enmime.Builder().AddAttachment([]byte("bar"), "ct", "fn")
	if a.Equals(b) {
		t.Error("Different AddAttachment(value) should not be equal")
	}

	a = enmime.Builder().AddAttachment([]byte("foo"), "ct", "fn")
	b = a.AddAttachment([]byte("bar"), "ct", "fn")
	b1 := b.AddAttachment([]byte("baz"), "ct", "fn")
	b2 := b.AddAttachment([]byte("bax"), "ct", "fn")
	if a.Equals(b) || b.Equals(b1) || b1.Equals(b2) {
		t.Error("AddAttachment() should not mutate receiver, failed")
	}

	want := "fake JPG data"
	name := "photo.jpg"
	a = enmime.Builder().
		Text([]byte("text")).
		HTML([]byte("html")).
		From("name", "foo").
		Subject("foo").
		ToAddrs(addrSlice).
		AddAttachment([]byte(want), "image/jpeg", name)
	root, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	p := root.DepthMatchFirst(func(p *enmime.Part) bool { return p.FileName == name })
	if p == nil {
		t.Fatalf("Did not find a %q part", name)
	}
	got := string(p.Content)
	if got != want {
		t.Errorf("Content: %q, want: %q", got, want)
	}

	// Check structure
	wantTypes := []string{
		"multipart/mixed",
		"multipart/alternative",
		"text/plain",
		"text/html",
		"image/jpeg",
	}
	gotParts := root.DepthMatchAll(func(p *enmime.Part) bool { return true })
	gotTypes := make([]string, 0)
	for _, p := range gotParts {
		gotTypes = append(gotTypes, p.ContentType)
	}
	test.DiffSlices(t, gotTypes, wantTypes)
}

func TestBuilderAddInline(t *testing.T) {
	a := enmime.Builder().AddInline([]byte("same"), "ct", "fn", "cid")
	b := enmime.Builder().AddInline([]byte("same"), "ct", "fn", "cid")
	if !a.Equals(b) {
		t.Error("Same AddInline(value) should be equal")
	}

	a = enmime.Builder().AddInline([]byte("foo"), "ct", "fn", "cid")
	b = enmime.Builder().AddInline([]byte("bar"), "ct", "fn", "cid")
	if a.Equals(b) {
		t.Error("Different AddInline(value) should not be equal")
	}

	a = enmime.Builder().AddInline([]byte("foo"), "ct", "fn", "cid")
	b = a.AddInline([]byte("bar"), "ct", "fn", "cid")
	b1 := b.AddInline([]byte("baz"), "ct", "fn", "cid")
	b2 := b.AddInline([]byte("bax"), "ct", "fn", "cid")
	if a.Equals(b) || b.Equals(b1) || b1.Equals(b2) {
		t.Error("AddInline() should not mutate receiver, failed")
	}

	want := "fake JPG data"
	name := "photo.jpg"
	cid := "<mycid>"
	a = enmime.Builder().
		Text([]byte("text")).
		HTML([]byte("html")).
		From("name", "foo").
		Subject("foo").
		ToAddrs(addrSlice).
		AddInline([]byte(want), "image/jpeg", name, cid)
	root, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	p := root.DepthMatchFirst(func(p *enmime.Part) bool { return p.ContentID == cid })
	if p == nil {
		t.Fatalf("Did not find a %q part", cid)
	}
	got := string(p.Content)
	if got != want {
		t.Errorf("Content: %q, want: %q", got, want)
	}

	// Check structure
	wantTypes := []string{
		"multipart/related",
		"multipart/alternative",
		"text/plain",
		"text/html",
		"image/jpeg",
	}
	gotParts := root.DepthMatchAll(func(p *enmime.Part) bool { return true })
	gotTypes := make([]string, 0)
	for _, p := range gotParts {
		gotTypes = append(gotTypes, p.ContentType)
	}
	test.DiffSlices(t, gotTypes, wantTypes)
}

func TestBuilderFullStructure(t *testing.T) {
	a := enmime.Builder().
		Text([]byte("text")).
		HTML([]byte("html")).
		From("name", "foo").
		Subject("foo").
		ToAddrs(addrSlice).
		AddAttachment([]byte("attach data"), "image/jpeg", "image.jpg").
		AddInline([]byte("inline data"), "image/png", "image.png", "")
	root, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}

	// Check structure via "parent > child" content types
	wantTypes := []string{
		" > multipart/mixed",
		"multipart/mixed > multipart/related",
		"multipart/related > multipart/alternative",
		"multipart/alternative > text/plain",
		"multipart/alternative > text/html",
		"multipart/related > image/png",
		"multipart/mixed > image/jpeg",
	}
	gotParts := root.DepthMatchAll(func(p *enmime.Part) bool { return true })
	gotTypes := make([]string, 0)
	for _, p := range gotParts {
		pct := ""
		if p.Parent != nil {
			pct = p.Parent.ContentType
		}
		gotTypes = append(gotTypes, pct+" > "+p.ContentType)
	}
	test.DiffSlices(t, gotTypes, wantTypes)
}

func TestHeader(t *testing.T) {
	a := enmime.Builder().Header("name", "same")
	b := enmime.Builder().Header("name", "same")
	if !a.Equals(b) {
		t.Error("Same Header(value) should be equal")
	}

	a = enmime.Builder().Header("name", "foo")
	b = enmime.Builder().Header("name", "bar")
	if a.Equals(b) {
		t.Error("Different Header(value) should not be equal")
	}

	a = enmime.Builder().Header("name", "foo")
	b = a.Header("name", "bar")
	if a.Equals(b) {
		t.Error("Header() should not mutate receiver, failed")
	}

	want := []string{"value one", "another value"}
	a = enmime.Builder().ToAddrs(addrSlice).From("name", "foo").Subject("foo")
	for _, s := range want {
		a = a.Header("X-Test", s)
	}
	p, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := p.Header["X-Test"]
	test.DiffSlices(t, got, want)
}
