package enmime_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/jhillyerd/enmime/v2"
	"github.com/stretchr/testify/assert"
)

// TestRandOption checks that different randomness modes behave as expected, relative to one another.
func TestRandOption(t *testing.T) {
	types := []ReproducibilityMode{ZeroSource, OneSource, DefaultSource, TimestampSource}
	for _, a := range types {
		for _, b := range types {
			ha, hb := buildEmail(t, a), buildEmail(t, b)
			if a == b && a.IsReproducible() {
				assert.Equal(t, ha, hb)
			} else {
				assert.NotEqual(t, ha, hb)
			}
		}
	}
}

type ReproducibilityMode int

const (
	ZeroSource ReproducibilityMode = iota
	OneSource
	DefaultSource
	TimestampSource
)

func (mode ReproducibilityMode) IsReproducible() bool {
	switch mode {
	case ZeroSource:
		return true
	case OneSource:
		return true
	case DefaultSource:
		return false
	case TimestampSource:
		return false
	default:
		panic(fmt.Errorf("illegal mode: %d", mode))
	}
}

func (mode ReproducibilityMode) String() string {
	switch mode {
	case ZeroSource:
		return "ZeroSource"
	case OneSource:
		return "OneSource"
	case DefaultSource:
		return "DefaultSource"
	case TimestampSource:
		return "TimestampSource"
	default:
		panic(fmt.Errorf("illegal mode: %d", mode))
	}
}

// buildEmail creates a string email, according to the given Reproducibilitymode.
func buildEmail(t *testing.T, mode ReproducibilityMode) string {
	t.Helper()
	var b enmime.MailBuilder
	switch mode {
	case ZeroSource:
		b = enmime.Builder().RandSeed(0)
	case OneSource:
		b = enmime.Builder().RandSeed(1)
	case DefaultSource:
		b = enmime.Builder()
	case TimestampSource:
		b = enmime.Builder().RandSeed(time.Now().UTC().UnixNano())
	default:
		panic(fmt.Errorf("illegal mode: %d", mode))
	}
	b = b.From("name", "same").To("anon", "anon@example.com").AddAttachment([]byte("testing"), "text/plain", "test.txt")
	p, err := b.Build()
	if err != nil {
		t.Fatalf("can't build email: %v", err)
	}
	w := new(bytes.Buffer)
	if err := p.Encode(w); err != nil {
		t.Fatalf("can't encode part: %v", err)
	}
	return w.String()
}
