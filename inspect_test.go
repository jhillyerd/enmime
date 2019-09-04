package enmime_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/mail"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/internal/test"
)

func TestDecodeHeaders(t *testing.T) {
	t.Run("rfc2047 sample", func(t *testing.T) {
		r := test.OpenTestData("mail", "qp-utf8-header.raw")
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Errorf("%+v", err)
		}
		h, err := enmime.DecodeHeaders(b)
		if err != nil {
			t.Errorf("%+v", err)
		}
		if !strings.Contains(h.Get("To"), "Mirosław Marczak") {
			t.Errorf("Error decoding RFC2047 header value")
		}
	})

	t.Run("rfc2047 recursive sample", func(t *testing.T) {
		r := test.OpenTestData("mail", "qp-utf8-header-recursed.raw")
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Errorf("%+v", err)
		}
		h, err := enmime.DecodeHeaders(b)
		if err != nil {
			t.Errorf("%+v", err)
		}
		if !strings.Contains(h.Get("From"), "WirelessCaller (203) 402-5984 WirelessCaller (203) 402-5984 WirelessCaller (203) 402-5984") {
			t.Errorf("Error decoding recursive RFC2047 header value")
		}
	})
}

func BenchmarkHumanHeadersOnly(b *testing.B) {
	r := test.OpenTestData("mail", "qp-utf8-header.raw")
	eml, err := ioutil.ReadAll(r)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		h, err := enmime.DecodeHeaders(eml)
		if err != nil {
			b.Fatal(err)
		}
		_, err = mail.ParseAddressList(h.Get("From"))
		if err != nil {
			b.Fatal(err)
		}
		_, err = mail.ParseAddressList(h.Get("To"))
		if err != nil {
			b.Fatal(err)
		}
		_ = h.Get("Subject")
	}
}

func BenchmarkReadEnvelope(b *testing.B) {
	r := test.OpenTestData("mail", "qp-utf8-header.raw")
	eml, err := ioutil.ReadAll(r)
	if err != nil {
		b.Fatal(err)
	}
	reusedReader := bytes.NewReader(eml)
	for i := 0; i < b.N; i++ {
		env, err := enmime.ReadEnvelope(reusedReader)
		if err != nil {
			b.Fatal(err)
		}
		_, err = env.AddressList("From")
		if err != nil {
			b.Fatal(err)
		}
		_, err = env.AddressList("To")
		if err != nil {
			b.Fatal(err)
		}
		env.GetHeader("Subject")
		// reset reader for next run
		_, err = reusedReader.Seek(0, io.SeekStart)
	}
}
