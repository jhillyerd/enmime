package enmime_test

import (
	"bytes"
	"io"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/internal/test"
)

func TestDecodeRFC2047(t *testing.T) {
	t.Run("rfc2047 basic", func(t *testing.T) {
		s := enmime.DecodeRFC2047("=?UTF-8?Q?Miros=C5=82aw_Marczak?=")
		if s != "Mirosław Marczak" {
			t.Errorf("Wrong decoded result")
		}
	})

	t.Run("rfc2047 unknown", func(t *testing.T) {
		s := enmime.DecodeRFC2047("=?ABC-1?Q?FooBar?=")
		if s != "=?ABC-1?Q?FooBar?=" {
			t.Errorf("Expected unmodified result for unknown charset")
		}
	})

	t.Run("rfc2047 pass-through", func(t *testing.T) {
		s := enmime.DecodeRFC2047("Hello World")
		if s != "Hello World" {
			t.Errorf("Expected unmodified result")
		}
	})
}

func TestDecodeHeaders(t *testing.T) {
	t.Run("rfc2047 sample", func(t *testing.T) {
		r := test.OpenTestData("mail", "qp-utf8-header.raw")
		b, err := io.ReadAll(r)
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

	t.Run("no break between headers and content", func(t *testing.T) {
		r := test.OpenTestData("mail", "qp-utf8-header-no-break.raw")
		b, err := io.ReadAll(r)
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

	t.Run("textproto header read error", func(t *testing.T) {
		r := test.OpenTestData("low-quality", "bad-header-start.raw")
		b, err := io.ReadAll(r)
		if err != nil {
			t.Errorf("%+v", err)
		}
		_, err = enmime.DecodeHeaders(b)
		switch err.(type) {
		case textproto.ProtocolError:
			// carry on
		default:
			t.Fatalf("Did return expected error: %T:%+v", err, err)
		}
	})

	t.Run("rfc2047 recursive sample", func(t *testing.T) {
		r := test.OpenTestData("mail", "qp-utf8-header-recursed.raw")
		b, err := io.ReadAll(r)
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
	eml, err := io.ReadAll(r)
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
	eml, err := io.ReadAll(r)
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
		if err != nil {
			b.Fatal(err)
		}
	}
}
