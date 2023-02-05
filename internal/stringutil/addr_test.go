package stringutil_test

import (
	"net/mail"
	"testing"

	"github.com/jhillyerd/enmime/internal/stringutil"
)

func TestJoinAddressEmpty(t *testing.T) {
	got := stringutil.JoinAddress(make([]mail.Address, 0))
	if got != "" {
		t.Errorf("Empty list got: %q, wanted empty string", got)
	}
}

func TestJoinAddressSingle(t *testing.T) {
	input := []mail.Address{
		{Name: "", Address: "one@bar.com"},
	}
	want := "<one@bar.com>"
	got := stringutil.JoinAddress(input)
	if got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}

	input = []mail.Address{
		{Name: "one name", Address: "one@bar.com"},
	}
	want = `"one name" <one@bar.com>`
	got = stringutil.JoinAddress(input)
	if got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestJoinAddressMany(t *testing.T) {
	input := []mail.Address{
		{Name: "one", Address: "one@bar.com"},
		{Name: "", Address: "two@foo.com"},
		{Name: "three", Address: "three@baz.com"},
	}
	want := `"one" <one@bar.com>, <two@foo.com>, "three" <three@baz.com>`
	got := stringutil.JoinAddress(input)
	if got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestCommaDelimitedAddressLists(t *testing.T) {
	testData := []struct {
		have string
		want string
	}{
		{
			have: `"Joe @ Company" <joe@company.com> <other@company.com>`,
			want: `"Joe @ Company" <joe@company.com>, <other@company.com>`,
		},
		{
			have: `Joe Company <joe@company.com> <other@company.com>`,
			want: `Joe Company <joe@company.com>, <other@company.com>`,
		},
		{
			have: `Joe Company:Joey <joe@company.com> John <other@company.com>;`,
			want: `Joe Company:Joey <joe@company.com>, John <other@company.com>`,
		},
		{
			have: `Joe Company:Joey <joe@company.com> John <other@company.com>; Jimmy John <jimmy.john@company.com>`,
			want: `Joe Company:Joey <joe@company.com>, John <other@company.com>, Jimmy John <jimmy.john@company.com>`,
		},
		{
			have: `Joe Company <joe@company.com> John Company <other@company.com>`,
			want: `Joe Company <joe@company.com>, John Company <other@company.com>`,
		},
		{
			have: `Joe Company <joe@company.com>,John Company <other@company.com>`,
			want: `Joe Company <joe@company.com>,John Company <other@company.com>`,
		},
		{
			have: `joe@company.com other@company.com`,
			want: `joe@company.com, other@company.com`,
		},
		{
			have: `Jimmy John <jimmy.john@company.com> joe@company.com other@company.com`,
			want: `Jimmy John <jimmy.john@company.com>, joe@company.com, other@company.com`,
		},
		{
			have: `Jimmy John <jimmy.john@company.com> joe@company.com John Company <other@company.com>`,
			want: `Jimmy John <jimmy.john@company.com>, joe@company.com, John Company <other@company.com>`,
		},
		{
			have: `<boss@nil.test> "Giant; \"Big\" Box" <sysservices@example.net>`,
			want: `<boss@nil.test>, "Giant; \"Big\" Box" <sysservices@example.net>`,
		},
		{
			have: `A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;`,
			want: `A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>`,
		},
		{
			have: `A Group:Ed Jones <c@a.test> joe@where.test John <jdoe@one.test>;`,
			want: `A Group:Ed Jones <c@a.test>, joe@where.test, John <jdoe@one.test>`,
		},
		{
			have: "a@example.com;b@example.com",
			want: "a@example.com,b@example.com",
		},
		{
			have: `"Joe @ Company" <joe@company.com>;<other@company.com>`,
			want: `"Joe @ Company" <joe@company.com>,<other@company.com>`,
		},
	}
	for _, tt := range testData {
		t.Run(tt.have, func(t *testing.T) {
			v := stringutil.EnsureCommaDelimitedAddresses(tt.have)
			if tt.want != v {
				t.Errorf("got: %q\nwant: %q", v, tt.want)
			}
		})
	}
}
