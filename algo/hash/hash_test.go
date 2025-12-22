package hash

import (
	"testing"
)

func TestMemhash(t *testing.T) {
	for _, s := range []string{"", "a", "ab", "abc"} {
		t.Logf("memhash(%q) = %v", s, MemHashString(s))
	}
}
