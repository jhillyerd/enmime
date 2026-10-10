package enmime_test

import (
	"bytes"
	"io"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"

	"github.com/jhillyerd/enmime/v2"
	"github.com/jhillyerd/enmime/v2/internal/test"
	inttp "github.com/jhillyerd/enmime/v2/internal/textproto"
)

func TestDecodeRFC2047(t *testing.T) {
	t.Run("rfc2047 basic", func(t *testing.T) {
		s := enmime.DecodeRFC2047("=?UTF-8?Q?Miros=C5=82aw_Marczak?=")
		if s != "Mirosław Marczak" {
			t.Error("Wrong decoded result")
		}
	})

	t.Run("rfc2047 unknown", func(t *testing.T) {
		s := enmime.DecodeRFC2047("=?ABC-1?Q?FooBar?=")
		if s != "=?ABC-1?Q?FooBar?=" {
			t.Error("Expected unmodified result for unknown charset")
		}
	})

	t.Run("rfc2047 pass-through", func(t *testing.T) {
		s := enmime.DecodeRFC2047("Hello World")
		if s != "Hello World" {
			t.Error("Expected unmodified result")
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
			t.Error("Error decoding RFC2047 header value")
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
			t.Error("Error decoding RFC2047 header value")
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
		case inttp.ProtocolError:
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
			t.Error("Error decoding recursive RFC2047 header value")
		}
	})

	t.Run("returns stdlib net/textproto MIMEHeader with additional headers", func(t *testing.T) {
		raw := []byte("From: =?UTF-8?Q?Miros=C5=82aw_Marczak?= <marczak@example.com>\r\n" +
			"To: recipient@example.com\r\n" +
			"Subject: =?UTF-8?B?VGVzdCBzdWJqZWN0?=\r\n" +
			"Date: Mon, 1 Jan 2024 12:00:00 +0000\r\n" +
			"X-Custom-Header: first-value\r\n" +
			"X-Custom-Header: second-value\r\n" +
			"\r\nBody text\r\n")

		// Explicitly typed as stdlib textproto.MIMEHeader to verify return type.
		var h textproto.MIMEHeader
		var err error
		h, err = enmime.DecodeHeaders(raw, "x-custom-header", "x-missing-header")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := h.Get("from"); got != "Mirosław Marczak <marczak@example.com>" {
			t.Errorf("h.Get(%q) = %q, want %q", "from", got, "Mirosław Marczak <marczak@example.com>")
		}
		if got := h.Get("To"); got != "recipient@example.com" {
			t.Errorf("h.Get(%q) = %q, want %q", "To", got, "recipient@example.com")
		}
		if got := h.Get("subject"); got != "Test subject" {
			t.Errorf("h.Get(%q) = %q, want %q", "subject", got, "Test subject")
		}
		if got := h.Get("x-custom-header"); got != "first-value" {
			t.Errorf("h.Get(%q) = %q, want %q", "x-custom-header", got, "first-value")
		}
		vals := h.Values("X-Custom-Header")
		if len(vals) != 2 || vals[0] != "first-value" || vals[1] != "second-value" {
			t.Errorf("h.Values(%q) = %v, want [first-value second-value]", "X-Custom-Header", vals)
		}
		if got := h.Get("x-missing-header"); got != "" {
			t.Errorf("h.Get(%q) = %q, want empty", "x-missing-header", got)
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
