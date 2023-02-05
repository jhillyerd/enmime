package stringutil_test

import (
	"testing"

	"github.com/jhillyerd/enmime/internal/stringutil"
)

func TestWrapEmpty(t *testing.T) {
	b := stringutil.Wrap(80, "")
	got := string(b)

	if got != "" {
		t.Errorf(`got: %q, want: ""`, got)
	}
}

func TestWrapIdentityShort(t *testing.T) {
	want := "short string"
	b := stringutil.Wrap(15, want)
	got := string(b)

	if got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestWrapIdentityLong(t *testing.T) {
	want := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	b := stringutil.Wrap(5, want)
	got := string(b)

	if got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestWrap(t *testing.T) {
	testCases := []struct {
		input, want string
	}{
		{
			"one two three",
			"one\r\n two\r\n three",
		},
		{
			"a bb ccc dddd eeeee ffffff",
			"a bb\r\n ccc\r\n dddd\r\n eeeee\r\n ffffff",
		},
		{
			"aaaaaa bbbbb cccc ddd ee f",
			"aaaaaa\r\n bbbbb\r\n cccc\r\n ddd\r\n ee f",
		},
		{
			"1 3 5 1 3 5 1 3 5",
			"1 3 5\r\n 1 3 5\r\n 1 3 5",
		},
		{
			"55555 55555 55555",
			"55555\r\n 55555\r\n 55555",
		},
		{
			"666666 666666 666666",
			"666666\r\n 666666\r\n 666666",
		},
		{
			"7777777 7777777 7777777",
			"7777777\r\n 7777777\r\n 7777777",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			b := stringutil.Wrap(6, tc.input)
			got := string(b)
			if got != tc.want {
				t.Errorf("got: %q, want: %q", got, tc.want)
			}

		})
	}
}
