package rolling

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrdinal(t *testing.T) {
	r := NewTiming()
	assert.EqualValues(t, 0, r.Mean())

	var ordinalTests = []struct {
		length   int
		perc     float64
		expected int64
	}{
		{1, 0, 1},
		{2, 0, 1},
		{2, 50, 1},
		{2, 51, 2},
		{5, 30, 2},
		{5, 40, 2},
		{5, 50, 3},
		{11, 25, 3},
		{11, 50, 6},
		{11, 75, 9},
		{11, 100, 11},
	}

	for _, s := range ordinalTests {
		assert.Equal(t, s.expected, r.ordinal(s.length, s.perc))
	}

	r.Add(100 * time.Millisecond)
	time.Sleep(2 * time.Second)
	r.Add(200 * time.Millisecond)
	assert.EqualValues(t, 150, r.Mean())

	durations := []int{1, 1004, 1004, 1004, 1004, 1004, 1004, 1004, 1004, 1004, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1006, 1006, 1006, 1006, 1007, 1007, 1007, 1008, 1015}
	for _, d := range durations {
		r.Add(time.Duration(d) * time.Millisecond)
	}

	assert.EqualValues(t, 1, r.Percentile(0))
	assert.EqualValues(t, 1006, r.Percentile(75))
	assert.EqualValues(t, 1015, r.Percentile(99))
	assert.EqualValues(t, 1015, r.Percentile(100))
}
