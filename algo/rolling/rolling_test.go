package rolling_test

import (
	"testing"
	"time"

	"github.com/cocktail828/go-tools/algo/rolling"
	"github.com/stretchr/testify/assert"
)

func TestRate(t *testing.T) {
	r := rolling.NewRolling(5, 5)
	for i := int64(0); i < 13; i++ {
		r.Incrby(i*8, 100*(int(i/8)+1))
	}

	assert.EqualValues(t, 162.5, r.Rate(96, 8))
	assert.EqualValues(t, 162.5, r.Rate(96, 80))
	assert.EqualValues(t, 0, r.Rate(1000000, 8))
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
