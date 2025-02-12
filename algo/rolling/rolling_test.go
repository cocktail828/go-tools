package rolling_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/algo/rolling"
	"github.com/stretchr/testify/assert"
)

func TestRate(t *testing.T) {
	r := rolling.NewRolling(0, 0)
	for i := int64(0); i < 13; i++ {
		r.Incrby(i*rolling.MIN_COUNTER_SIZE, 100*int(i+1))
	}

	assert.EqualValues(t, 750, r.Rate(12*rolling.MIN_COUNTER_SIZE, 12))
	assert.EqualValues(t, 5, r.Rate(0, 20))
	assert.EqualValues(t, 0, r.Rate(1000000, 8))
}

func TestIncrExpire(t *testing.T) {
	r := rolling.NewRolling(0, 0)
	r.Incrby(0, 100)

	msec := rolling.MIN_COUNTER_NUM * rolling.MIN_COUNTER_SIZE
	r.Incrby(int64(msec), 23)
	cnt, win := r.Count(int64(msec), 1)
	assert.EqualValues(t, 23, cnt)
	assert.EqualValues(t, 1, win)
}

func BenchmarkConcurrency(b *testing.B) {
	r := rolling.NewRolling(0, 0)
	cnt := atomic.Int64{}
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.Incrby(0, 1)
			cnt.Add(1)
		}
	})

	v, _ := r.Count(0, 1)
	assert.EqualValues(b, v, cnt.Load())
}

func BenchmarkRolling(b *testing.B) {
	r := rolling.NewRolling(100, 100)

	msec := time.Now().UnixMilli()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.Rate(msec, 8)
		}
	})
}
