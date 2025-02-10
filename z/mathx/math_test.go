package mathx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNext2Power(t *testing.T) {
	for k, v := range map[int64]int64{
		-15: -8,
		-16: -16,
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
