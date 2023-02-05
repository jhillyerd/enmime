package stringutil_test

import (
	"testing"

	"github.com/jhillyerd/enmime/internal/stringutil"
)

func TestUUID(t *testing.T) {
	id1 := stringutil.UUID(nil)
	id2 := stringutil.UUID(nil)

	if id1 == id2 {
		t.Errorf("Random UUID should not equal another random UUID")
		t.Logf("id1: %q", id1)
		t.Logf("id2: %q", id2)
	}
}
