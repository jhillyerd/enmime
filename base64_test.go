package enmime

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64Cleaner(t *testing.T) {
	input := strings.NewReader("\tA B\r\nC")
	cleaner := NewBase64Cleaner(input)
	buf := new(bytes.Buffer)
	buf.ReadFrom(cleaner)

	assert.Equal(t, "ABC", buf.String())
}
