package stringutil

import (
	"testing"
)

func TestSplitQuoted(t *testing.T) {
	testCases := []struct {
		input string
		want  []string
	}{
		{
			input: "",
			want:  []string{""},
		},
		{
			input: ";",
			want:  []string{"", ""},
		},
		{
			input: "'",
			want:  []string{"'"},
		},
		{
			input: "a;b",
			want:  []string{"a", "b"},
		},
		{
			input: "a;b;",
			want:  []string{"a", "b", ""},
		},
		{
			input: "a;'b;c';d",
			want:  []string{"a", "'b;c'", "d"},
		},
		{
			input: "a;'b;c;d",
			want:  []string{"a", "'b;c;d"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := SplitQuoted(tc.input, ';', '\'')

			t.Logf("\ngot : %q\nwant: %q\n", got, tc.want)

			if len(got) != len(tc.want) {
				t.Errorf("got len %v, want len %v", len(got), len(tc.want))
				return
			}

			for i, g := range got {
				if g != tc.want[i] {
					t.Errorf("Element %v differs", i)
				}
			}
		})
	}
}
