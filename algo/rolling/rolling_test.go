package rolling

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/timex"
	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	timex.SetTime(func() int64 { return time.Minute.Nanoseconds() })
	r := NewRolling(128)
	r.At(timex.UnixNano()).DualIncrBy(1, 0)
	c, _, w := r.DualCount(10)

	assert.EqualValues(t, 1, c)
	assert.EqualValues(t, 1, w)

	c, _, w = r.At(timex.UnixNano()).DualCount(10)
	assert.EqualValues(t, 1, c)
	assert.EqualValues(t, 1, w)
}

func TestQPS(t *testing.T) {
	r := NewRolling(128)
	for i := int64(0); i < 13; i++ {
		timex.SetTime(func() int64 { return i * ROLLING_PRECISION })
		r.IncrBy(100 * int(i+1))
	}

	timex.SetTime(func() int64 { return 12 * ROLLING_PRECISION })
	assert.EqualValues(t, 5859.375, r.QPS(12))

	timex.SetTime(func() int64 { return 0 })
	assert.EqualValues(t, 781.25, r.QPS(1))
	assert.EqualValues(t, 39.0625, r.QPS(20))

	timex.SetTime(func() int64 { return 1000000 * 1e6 })
	assert.EqualValues(t, 0, r.QPS(8))
}

func TestIncrExpire(t *testing.T) {
	r := NewRolling(0)
	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(100)

	timex.SetTime(func() int64 { return ROLLING_MIN_COUNTER * ROLLING_PRECISION })
	r.IncrBy(23)
	cnt, win := r.Count(1)
	assert.EqualValues(t, 23, cnt)
	assert.EqualValues(t, 1, win)
}

func TestGettime(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })
	assert.EqualValues(t, 0, timex.UnixNano())

	timex.SetTime(func() int64 { return 1000 })
	assert.EqualValues(t, 1000, timex.UnixNano())
}

func BenchmarkConcurrency(b *testing.B) {
	r := NewRolling(0)
	cnt := atomic.Int64{}
	timex.SetTime(func() int64 { return 0 })
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.IncrBy(1)
			cnt.Add(1)
		}
	})

	v, _ := r.Count(1)
	assert.EqualValues(b, v, cnt.Load())
}

func BenchmarkRolling(b *testing.B) {
	r := NewRolling(100)

	timex.SetTime(func() int64 { return 0 })
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.QPS(8)
		}
	})
}
