package enmime_test

import (
	"io/ioutil"
	"net/mail"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/internal/test"
)

func TestHumanHeadersOnly(t *testing.T) {
	t.Run("rfc2047 sample", func(t *testing.T) {
		r := test.OpenTestData("mail", "qp-utf8-header.raw")
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Errorf("%+v", err)
		}
		humanHeaders, err := enmime.HumanHeadersOnly(b)
		if err != nil {
			t.Errorf("%+v", err)
		}
		if !strings.Contains(humanHeaders["To"], "Mirosław Marczak") {
			t.Errorf("Error decoding RFC2047 header value")
		}
	})

	t.Run("rfc2047 recursive sample", func(t *testing.T) {
		r := test.OpenTestData("mail", "qp-utf8-header-recursed.raw")
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Errorf("%+v", err)
		}
		humanHeaders, err := enmime.HumanHeadersOnly(b)
		if err != nil {
			t.Errorf("%+v", err)
		}
		if !strings.Contains(humanHeaders["From"], "WirelessCaller (203) 402-5984 WirelessCaller (203) 402-5984 WirelessCaller (203) 402-5984") {
			t.Errorf("Error decoding recursive RFC2047 header value")
		}
	})
}

func BenchmarkHumanHeadersOnly(b *testing.B) {
	r := test.OpenTestData("mail", "qp-utf8-header.raw")
	test, _ := ioutil.ReadAll(r)
	for i := 0; i < b.N; i++ {
		h, _ := enmime.HumanHeadersOnly(test)
		mail.ParseAddressList(h["From"])
		mail.ParseAddressList(h["To"])
		_ = h["Subject"]
	}
}

func BenchmarkReadEnvelope(b *testing.B) {
	r := test.OpenTestData("mail", "qp-utf8-header.raw")
	for i := 0; i < b.N; i++ {
		env, _ := enmime.ReadEnvelope(r)
		env.AddressList("From")
		env.AddressList("To")
		env.GetHeader("Subject")
	}
}
