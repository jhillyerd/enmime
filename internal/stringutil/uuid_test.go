package stringutil_test

import (
	"testing"

	"github.com/jhillyerd/enmime/internal/stringutil"
	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	id1 := stringutil.UUID(nil)
	id2 := stringutil.UUID(nil)

	assert.NotEqual(t, id1, id2, "Random UUID should not equal another random UUID")
}
