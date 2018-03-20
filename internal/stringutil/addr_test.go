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
