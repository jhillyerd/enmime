package stringutil

import "testing"

func TestFindQuoted(t *testing.T) {
	findRune := ';'
	quoteRune := '"'
	testCases := []struct {
		input string
		want  []int
	}{
		{
			input: ``,
			want:  []int{},
		},
		{
			input: `;`,
			want:  []int{0},
		},
		{
			input: `\;`, // not escaped
			want:  []int{1},
		},
		{
			input: `a;b`, // single
			want:  []int{1},
		},
		{
			input: `a\;b;c`, // not escaped, multiple
			want:  []int{2, 4},
		},
		{
			input: `\`, // just escape
			want:  []int{},
		},
		{
			input: `\\;`, // escaped escape
			want:  []int{2},
		},
		{
			input: `"`, // just quote
			want:  []int{},
		},
		{
			input: `a";b"`, // inside quotes - ignored
			want:  []int{},
		},
		{
			input: `"a"";b"`, // inside quotes - ignored
			want:  []int{},
		},
		{
			input: `a\";b"`, // escaped quote - ignore quote
			want:  []int{3},
		},
		{
			input: `"a;b;c;d`, // unterminated quote at the beginning - ignore
			want:  []int{2, 4, 6},
		},
		{
			input: `a"`, // unterminated quote at the end - ignore
			want:  []int{},
		},
		{
			input: `ab";c`, // unterminated quote at the middle - ignore
			want:  []int{3},
		},
		{
			input: `"a;b""c;d`, // unterminated quote with properly terminated quote
			want:  []int{7},
		},
		{
			input: `a"b\";\"c";d`, // inside escaped quotes which should be ignored
			want:  []int{10},
		},
		{
			input: `a;"b";""`,
			want:  []int{1, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := FindQuoted(tc.input, findRune, quoteRune)

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
