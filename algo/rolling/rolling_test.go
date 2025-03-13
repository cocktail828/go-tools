package rolling

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRate(t *testing.T) {
	r := NewRolling(0, 0)
	for i := int64(0); i < 13; i++ {
		SetTime(func() int64 { return i * MIN_COUNTER_SIZE * 1e6 })
		r.Incrby(100 * int(i+1))
	}

	SetTime(func() int64 { return 12 * MIN_COUNTER_SIZE * 1e6 })
	assert.EqualValues(t, 750, r.Rate(12))
	SetTime(func() int64 { return 0 })
	assert.EqualValues(t, 5, r.Rate(20))
	SetTime(func() int64 { return 1000000 * 1e6 })
	assert.EqualValues(t, 0, r.Rate(8))
}

func TestIncrExpire(t *testing.T) {
	r := NewRolling(0, 0)
	SetTime(func() int64 { return 0 })
	r.Incrby(100)

	SetTime(func() int64 { return MIN_COUNTER_NUM * MIN_COUNTER_SIZE * 1e6 })
	r.Incrby(23)
	cnt, win := r.Count(1)
	assert.EqualValues(t, 23, cnt)
	assert.EqualValues(t, 1, win)
}

func TestGettime(t *testing.T) {
	SetTime(func() int64 { return 0 })
	assert.EqualValues(t, 0, unixNano())

	SetTime(func() int64 { return 1000 })
	assert.EqualValues(t, 1000, unixNano())
}

func BenchmarkConcurrency(b *testing.B) {
	r := NewRolling(0, 0)
	cnt := atomic.Int64{}
	SetTime(func() int64 { return 0 })
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.Incrby(1)
			cnt.Add(1)
		}
	})

	v, _ := r.Count(1)
	assert.EqualValues(b, v, cnt.Load())
}

func BenchmarkRolling(b *testing.B) {
	r := NewRolling(100, 100)

	SetTime(func() int64 { return 0 })
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.Rate(8)
		}
	})
}
