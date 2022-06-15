package stringutil

import (
	"testing"
)

func TestSplitUnquoted(t *testing.T) {
	testCases := []struct {
		input string
		want  []string
	}{
		// All tests split on ; and treat " as quoting character.
		{
			input: ``,
			want:  []string{``},
		},
		{
			input: `;`,
			want:  []string{``, ``},
		},
		{
			input: `"`,
			want:  []string{`"`},
		},
		{
			input: `a;b`,
			want:  []string{`a`, `b`},
		},
		{
			input: `a;b;`,
			want:  []string{`a`, `b`, ``},
		},
		{
			input: `a;b;c`,
			want:  []string{`a`, `b`, `c`},
		},
		{
			// Separators are ignored within quoted-runs.
			input: `a;"b;c";d`,
			want:  []string{`a`, `"b;c"`, `d`},
		},
		{
			// Unterminated quoted-run will cause quotes to be ignored from the start of the string.
			input: `"a;b;c;d`,
			want:  []string{`"a`, `b`, `c`, `d`},
		},
		{
			input: `"a;b";"c;d`,
			want:  []string{`"a;b"`, `"c`, `d`},
		},
		{
			input: `a;"b\";\"c";d`,
			want:  []string{`a`, `"b\";\"c"`, `d`},
		},
		{
			input: `a;"b";""`,
			want:  []string{`a`, `"b"`, `""`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := SplitUnquoted(tc.input, ';', '"')

			if len(got) != len(tc.want) {
				t.Errorf("got len %v, want len %v", len(got), len(tc.want))
				return
			}

			for i, g := range got {
				if g != tc.want[i] {
					t.Errorf("element %v differs: got %q, want %q", i, g, tc.want[i])
				}
			}
		})
	}
}

func TestSplitAfterUnquoted(t *testing.T) {
	testCases := []struct {
		input string
		want  []string
	}{
		// All tests split on ; and treat " as quoting character.
		{
			input: ``,
			want:  []string{``},
		},
		{
			input: `;`,
			want:  []string{`;`, ``},
		},
		{
			input: `"`,
			want:  []string{`"`},
		},
		{
			input: `a;b`,
			want:  []string{`a;`, `b`},
		},
		{
			input: `a;b;`,
			want:  []string{`a;`, `b;`, ``},
		},
		{
			input: `a;b;c`,
			want:  []string{`a;`, `b;`, `c`},
		},
		{
			// Separators are ignored within quoted-runs.
			input: `a;"b;c";d`,
			want:  []string{`a;`, `"b;c";`, `d`},
		},
		{
			// Unterminated quoted-run will cause quotes to be ignored from the start of the string.
			input: `"a;b;c;d`,
			want:  []string{`"a;`, `b;`, `c;`, `d`},
		},
		{
			input: `"a;b";"c;d`,
			want:  []string{`"a;b";`, `"c;`, `d`},
		},
		{
			input: `a;"b\";\"c";d`,
			want:  []string{`a;`, `"b\";\"c";`, `d`},
		},
		{
			input: `a;b\;c`,
			want:  []string{`a;`, `b\;`, `c`},
		},
		{
			input: `a;"b";""`,
			want:  []string{`a;`, `"b";`, `""`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := SplitAfterUnquoted(tc.input, ';', '"')

			if len(got) != len(tc.want) {
				t.Errorf("got len %v, want len %v", len(got), len(tc.want))
				return
			}

			for i, g := range got {
				if g != tc.want[i] {
					t.Errorf("element %v differs: got %q, want %q", i, g, tc.want[i])
				}
			}
		})
	}
}
