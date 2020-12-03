package stringutil

import (
	"testing"
)

func TestSplitQuoted(t *testing.T) {
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
			// Unterminated quoted-run will cause quotes to be ignored from the start of the string.
			input: `"a;b";"c;d`,
			want:  []string{`"a`, `b"`, `"c`, `d`},
		},
		{
			// Quotes must be escaped via RFC2047 encoding, not just a backslash.
			// b through c below must not be treated as a single quoted-run.
			input: `a;"b\";\"c";d`,
			want:  []string{`a`, `"b\"`, `\"c"`, `d`},
		},
		{
			input: `a;b\;c`,
			want:  []string{`a`, `b\`, `c`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := SplitQuoted(tc.input, ';', '"')

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
