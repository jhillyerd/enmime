package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jhillyerd/enmime"
)

func TestHelperComparePartsEqual(t *testing.T) {
	testCases := []struct {
		name string
		part *enmime.Part
	}{
		{"nil", nil},
		{"empty", &enmime.Part{}},
		{"Parent", &enmime.Part{Parent: &enmime.Part{}}},
		{"FirstChild", &enmime.Part{FirstChild: &enmime.Part{}}},
		{"NextSibling", &enmime.Part{NextSibling: &enmime.Part{}}},
		{"ContentType", &enmime.Part{ContentType: "such/wow"}},
		{"Disposition", &enmime.Part{Disposition: "irritable"}},
		{"FileName", &enmime.Part{FileName: "readme.txt"}},
		{"Charset", &enmime.Part{Charset: "utf-7.999"}},
		{"PartID", &enmime.Part{PartID: "0.1"}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockt := &testing.T{}
			if !ComparePart(mockt, tc.part, tc.part) {
				t.Errorf("Got false while comparing a Part %v to itself: %+v", tc.name, tc.part)
			}
			if mockt.Failed() {
				t.Errorf("Helper failed test for %q, should have been successful", tc.name)
			}
		})
	}
}

// TestHelperComparePartsInequal tests compareParts with differing Parts
func TestHelperComparePartsInequal(t *testing.T) {
	testCases := []struct {
		name string
		a, b *enmime.Part
	}{
		{
			name: "nil",
			a:    nil,
			b:    &enmime.Part{},
		},
		{
			name: "Parent",
			a:    &enmime.Part{},
			b:    &enmime.Part{Parent: &enmime.Part{}},
		},
		{
			name: "FirstChild",
			a:    &enmime.Part{},
			b:    &enmime.Part{FirstChild: &enmime.Part{}},
		},
		{
			name: "NextSibling",
			a:    &enmime.Part{},
			b:    &enmime.Part{NextSibling: &enmime.Part{}},
		},
		{
			name: "ContentType",
			a:    &enmime.Part{ContentType: "text/plain"},
			b:    &enmime.Part{ContentType: "text/html"},
		},
		{
			name: "Disposition",
			a:    &enmime.Part{Disposition: "happy"},
			b:    &enmime.Part{Disposition: "sad"},
		},
		{
			name: "FileName",
			a:    &enmime.Part{FileName: "foo.gif"},
			b:    &enmime.Part{FileName: "bar.jpg"},
		},
		{
			name: "Charset",
			a:    &enmime.Part{Charset: "foo"},
			b:    &enmime.Part{Charset: "bar"},
		},
		{
			name: "PartID",
			a:    &enmime.Part{PartID: "0"},
			b:    &enmime.Part{PartID: "1.1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockt := &testing.T{}
			if ComparePart(mockt, tc.a, tc.b) {
				t.Errorf(
					"Got true while comparing inequal Parts (%v):\n"+
						"Part A: %+v\nPart B: %+v", tc.name, tc.a, tc.b)
			}
			if tc.name != "" && !mockt.Failed() {
				t.Errorf("Mock test succeeded for %s, should have failed", tc.name)
			}
		})
	}
}

// TestOpenTestDataPanic verifies that this function will panic as predicted
func TestOpenTestDataPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("OpenTestData did not panic")
		}
	}()
	_ = OpenTestData("invalidDir", "invalidFile")
}

// TestOpenTestData ensures that the returned io.Reader has the correct underlying type, that
// the file descriptor referenced is a directory and that we have permission to read it
func TestOpenTestData(t *testing.T) {
	// this will open a handle to the "testdata" directory
	r := OpenTestData("", "")
	if r == nil {
		t.Error("The returned io.Reader should not be nil")
	}

	osFilePtr, ok := r.(*os.File)
	if !ok {
		t.Errorf("Underlying type should be *os.File, but got %T instead", r)
	}

	info, err := osFilePtr.Stat()
	if err != nil {
		t.Error("We should have read permission for \"testdata\" directory")
	}

	if !info.IsDir() {
		t.Error("File descriptor labeled \"testdata\" should be a directory")
	}
}

// TestContentContainsString checks if the string contains a provided sub-string
func TestContentContainsString(t *testing.T) {
	// Success
	ContentContainsString(t, []byte("someString"), "some")
	// Failure
	ContentContainsString(&testing.T{}, []byte("someString"), "nope")
}

// TestContentEqualsString checks if the strings are equal
func TestContentEqualsString(t *testing.T) {
	// Success
	ContentEqualsString(t, []byte("someString"), "someString")
	// Failure
	ContentEqualsString(&testing.T{}, []byte("someString"), "nope")
}

// TestContentEqualsBytes checks if the slices of bytes are equal
func TestContentEqualsBytes(t *testing.T) {
	// Success
	ContentEqualsBytes(t, []byte("someString"), []byte("someString"))
	// Failure
	ContentEqualsBytes(&testing.T{}, []byte("someString"), []byte("nope"))
}

// TestCompareEnvelope checks all publicly accessible members of an envelope for differences
func TestCompareEnvelope(t *testing.T) {
	fileA, err := os.Open(filepath.Join("..", "..", "testdata", "mail", "attachment.raw"))
	if err != nil {
		t.Error(err)
	}
	envelopeA, err := enmime.ReadEnvelope(fileA)
	if err != nil {
		t.Error(err)
	}

	// Success
	success := CompareEnvelope(t, envelopeA, envelopeA)
	if !success {
		t.Error("Same file should have identical envelopes")
	}

	// Success on "want" and "got" nil
	success = CompareEnvelope(t, nil, nil)
	if !success {
		t.Error("Comparing nil to nil should result in true")
	}

	// Fail on "got" nil
	success = CompareEnvelope(&testing.T{}, nil, envelopeA)
	if success {
		t.Error("Got is nil, envelopeA should not be the same")
	}

	// Fail on "want" nil
	success = CompareEnvelope(&testing.T{}, envelopeA, nil)
	if success {
		t.Error("Want is nil, envelopeA should not be the same")
	}

	// Fail on root Part mismatch nil
	var envelopeB enmime.Envelope
	// copy the bytes and not the pointer
	envelopeB = *envelopeA
	envelopeB.Root = nil
	success = CompareEnvelope(&testing.T{}, envelopeA, &envelopeB)
	if success {
		t.Error("Envelope Root parts should not be the same")
	}
	envelopeB.Root = envelopeA.Root

	// Fail on Text mismatch
	envelopeB.Text = "mismatch"
	success = CompareEnvelope(&testing.T{}, envelopeA, &envelopeB)
	if success {
		t.Error("Envelope Text parts should not be the same")
	}
	envelopeB.Text = envelopeA.Text

	// Fail on HTML mismatch
	envelopeB.HTML = "mismatch"
	success = CompareEnvelope(&testing.T{}, envelopeA, &envelopeB)
	if success {
		t.Error("Envelope HTML parts should not be the same")
	}
	envelopeB.HTML = envelopeA.HTML

	// Fail on Attachment count mismatch
	envelopeB.Attachments = append(envelopeB.Attachments, &enmime.Part{})
	success = CompareEnvelope(&testing.T{}, envelopeA, &envelopeB)
	if success {
		t.Error("Envelope Attachment slices should not be the same")
	}
	envelopeB.Attachments = envelopeA.Attachments

	// Fail on Inlines count mismatch
	envelopeB.Inlines = append(envelopeB.Inlines, &enmime.Part{})
	success = CompareEnvelope(&testing.T{}, envelopeA, &envelopeB)
	if success {
		t.Error("Envelope Inlines slices should not be the same")
	}
	envelopeB.Inlines = envelopeA.Inlines

	// Fail on OtherParts count mismatch
	envelopeB.OtherParts = append(envelopeB.OtherParts, &enmime.Part{})
	success = CompareEnvelope(&testing.T{}, envelopeA, &envelopeB)
	if success {
		t.Error("Envelope OtherParts slices should not be the same")
	}
	envelopeB.OtherParts = envelopeA.OtherParts

	// Fail on Errors count mismatch
	envelopeB.Errors = append(envelopeB.Errors, &enmime.Error{})
	success = CompareEnvelope(&testing.T{}, envelopeA, &envelopeB)
	if success {
		t.Error("Envelope Errors slices should not be the same")
	}
	envelopeB.Errors = envelopeA.Errors
}
