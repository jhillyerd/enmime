package enmime_test

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/jhillyerd/enmime"
)

func TestRandOption(t *testing.T) {
	if hash(t, NewZeroSource) != hash(t, NewZeroSource) {
		t.Fatalf("hashes of email buffers differ")
	}
	if hash(t, Default) == hash(t, Default) {
		t.Fatalf("hashes of email buffers should differ")
	}
	if hash(t, Timestamp) == hash(t, Timestamp) {
		t.Fatalf("hashes of email buffers should differ")
	}
}

type Reproducibility int

const (
	NewZeroSource Reproducibility = iota
	Default
	Timestamp
)

func hash(t *testing.T, mode Reproducibility) string {
	var b enmime.MailBuilder
	switch mode {
	case NewZeroSource:
		b = enmime.Builder(enmime.RandBuilderOption(rand.New(rand.NewSource(0))))
	case Default:
		b = enmime.Builder()
	case Timestamp:
		b = enmime.Builder(enmime.RandBuilderOption(rand.New(rand.NewSource(time.Now().UTC().UnixNano()))))
	default:
		panic(fmt.Errorf("illegal mode: %d", mode))
	}
	b = b.From("name", "same").To("anon", "anon@example.com").AddAttachment([]byte("testing"), "plain/text", "test.txt")
	p, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	w := new(bytes.Buffer)
	if err := p.Encode(w); err != nil {
		t.Fatal(err)
	}
	h := md5.New()
	h.Write(w.Bytes())
	return fmt.Sprintf("0x%x", h.Sum(nil))
}
