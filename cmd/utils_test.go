package cmd

import (
	"bytes"
	"github.com/jhillyerd/enmime"
	"testing"
)

func TestFormatPartNil(t *testing.T) {
	buf := new(bytes.Buffer)
	FormatPart(buf, nil, "")
	got := buf.String()
	want := ""
	if got != want {
		t.Errorf("FormatPart(nil) == %q, want: %q", got, want)
	}
}

func TestFormatPartEmpty(t *testing.T) {
	buf := new(bytes.Buffer)
	FormatPart(buf, &enmime.Part{}, "")
	got := buf.String()
	want := "MISSING TYPE\n"
	if got != want {
		t.Errorf("FormatPart(nil) == %q, want: %q", got, want)
	}
}

func TestFormatPartMulti(t *testing.T) {
	buf := new(bytes.Buffer)

	// Build part tree
	root := &enmime.Part{
		ContentType: "multipart/alternative",
	}
	lev1 := &enmime.Part{
		ContentType: "text/plain",
		Parent:      root,
		NextSibling: &enmime.Part{
			ContentType: "text/html",
			Parent:      root,
			NextSibling: &enmime.Part{
				ContentType: "multipart/mixed",
				Parent:      root,
			},
		},
	}
	root.FirstChild = lev1
	lev2 := &enmime.Part{
		ContentType: "image/png",
		Disposition: "inline",
		FileName:    "test.png",
		Parent:      lev1,
		NextSibling: &enmime.Part{
			ContentType: "image/jpeg",
			Disposition: "attachment",
			FileName:    "test.jpg",
			Parent:      lev1,
		},
	}
	lev1.NextSibling.NextSibling.FirstChild = lev2

	// Setup an error
	lev1.Errors = []enmime.Error{
		{Name: "Test Error", Detail: "None", Severe: false},
	}

	// Desired output lines
	lines := []string{
		"multipart/alternative",
		"|-- text/plain (errors: 1)",
		"|-- text/html",
		"`-- multipart/mixed",
		"    |-- image/png, disposition: inline, filename: \"test.png\"",
		"    `-- image/jpeg, disposition: attachment, filename: \"test.jpg\"",
	}

	FormatPart(buf, root, "")

	for i, want := range lines {
		got, err := buf.ReadString('\n')
		if err != nil {
			t.Fatalf("Error on line %v: %v", i+1, err)
		}
		// Drop \n
		got = got[:len(got)-1]
		if got != want {
			t.Errorf("Line %v got: %q, want: %q", i+1, got, want)
		}
	}
}
