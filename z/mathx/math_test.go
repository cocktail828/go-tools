package mathx

import (
	"testing"
	"time"

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

func TestBase62(t *testing.T) {
	n := time.Now().UnixNano()
	assert.Equal(t, n, FromBase62(ToBase62(n)))
}

func TestCeilOf(t *testing.T) {
	tests := []struct {
		num, base, expt int64
	}{
		{100, 10, 100},
		{101, 10, 100},
		{102, 10, 100},
		{103, 10, 100},
		{104, 10, 100},
		{105, 10, 100},
		{106, 10, 100},
		{107, 10, 100},
		{108, 10, 100},
		{109, 10, 100},
		{110, 10, 110},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expt, CeilOf(tt.num, tt.base))
	}
}
