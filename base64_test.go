package enmime

import (
	"bytes"
	"github.com/stretchrcom/testify/assert"
	"strings"
	"testing"
)

func TestBase64Cleaner(t *testing.T) {
	input := strings.NewReader("\tA B\r\nC")
	cleaner := NewBase64Cleaner(input)
	buf := new(bytes.Buffer)
	buf.ReadFrom(cleaner)

	assert.Equal(t, buf.String(), "ABC")
}
