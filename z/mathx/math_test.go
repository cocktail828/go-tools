package mathx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNext2Power(t *testing.T) {
	for k, v := range map[int64]int64{
		0:   0,
		1:   1,
		2:   2,
		3:   4,
		4:   4,
		5:   8,
		6:   8,
		7:   8,
		8:   8,
		15:  16,
		125: 128,
	} {
		assert.Equalf(t, int64(v), Next2Power(k), "k=%v, expt=%v", k, v)
	}
}

func TestNumOfOnes(t *testing.T) {
	for k, v := range map[int64]int{
		-0b1111:     4,
		-0b11110011: 6,
		0:           0,
		1:           1,
		2:           1,
		4:           1,
	} {
		assert.Equalf(t, v, NumOfOnes(k), "k=%v, expt=%v", k, v)
	}
}

func TestCeilOf(t *testing.T) {
	tests := []struct {
		num, base, floor, ceil int64
	}{
		{100, 10, 100, 100},
		{101, 10, 100, 110},
		{102, 10, 100, 110},
		{103, 10, 100, 110},
		{104, 10, 100, 110},
		{105, 10, 100, 110},
		{106, 10, 100, 110},
		{107, 10, 100, 110},
		{108, 10, 100, 110},
		{109, 10, 100, 110},
		{110, 10, 110, 110},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.floor, Floor(tt.num, tt.base), "num=%v, base=%v", tt.num, tt.base)
		assert.Equal(t, tt.ceil, Ceil(tt.num, tt.base), "num=%v, base=%v", tt.num, tt.base)
	}
}

func TestMemhash(t *testing.T) {
	for _, s := range []string{"", "a", "ab", "abc"} {
		t.Logf("memhash(%s) = %v", s, MemHashString(s))
	}
}
