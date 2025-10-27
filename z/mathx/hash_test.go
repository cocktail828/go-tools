package mathx

import (
	"testing"
)

func TestMemhash(t *testing.T) {
	for _, s := range []string{"", "a", "ab", "abc"} {
		t.Logf("memhash(%q) = %v", s, MemHashString(s))
	}
}

func TestMurmurHash3_32(t *testing.T) {
	for _, s := range []string{"", "a", "ab", "abc"} {
		t.Logf("murmurhash3_32(%q) = %v", s, MurmurHash3_32([]byte(s), 0))
	}
}

func TestMurmurHash3_64(t *testing.T) {
	for _, s := range []string{"", "a", "ab", "abc"} {
		t.Logf("murmurhash3_64(%q) = %v", s, MurmurHash3_64([]byte(s), 0))
	}
}
