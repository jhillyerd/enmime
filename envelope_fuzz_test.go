//go:build go1.18

package enmime

import (
	"bytes"
	"testing"
)

func FuzzReadEnvelope(f *testing.F) {
	f.Add([]byte(`Date: Mon, 23 Jun 2015 11:40:36 -0400
From: Gopher <from@example.com>
To: Example User <user@example.com>
Subject: Gophers at Gophercon

Message body


`))
	f.Fuzz(func(t *testing.T, b []byte) {
		envelope, err := ReadEnvelope(bytes.NewReader(b))
		if envelope != nil && err != nil {
			t.Error("envelope and error are not nil")
		}
	})
}
