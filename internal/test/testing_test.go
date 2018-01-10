package test

import (
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
