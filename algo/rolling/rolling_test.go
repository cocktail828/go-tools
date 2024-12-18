package rolling

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRate(t *testing.T) {
	r := Rolling{}
	for i := 0; i < SlotPerBucket; i++ {
		r.buckets[0][i].Add(100)
	}
	for i := 0; i < SlotPerBucket; i++ {
		r.buckets[1][i].Add(200)
	}
	for i := 0; i < SlotPerBucket; i++ {
		r.buckets[2][i].Add(300)
	}
	assert.EqualValues(t, 200, r.rate(64, 110, 1))
	assert.EqualValues(t, 1200, r.rate(65, 110, 1))
	assert.EqualValues(t, 2200, r.rate(66, 110, 1))
}

func TestReset(t *testing.T) {
	r := Rolling{}
	for i := 0; i < BucketNum; i++ {
		for j := 0; j < SlotPerBucket; j++ {
			r.buckets[i][j].Add(100)
		}
	}

	r.SafeReset(0)
	for i := 0; i < BucketNum; i++ {
		for j := 0; j < SlotPerBucket; j++ {
			// fmt.Println(i, j, r.buckets[i][j].Load())
		}
	}
}

func BenchmarkRolling(b *testing.B) {
	r := Rolling{}
	r.IncrBy(1000)
	time.Sleep(time.Second)
	r.IncrBy(2000)
	time.Sleep(time.Second)
	r.IncrBy(3000)
	time.Sleep(time.Second)

	sec, nsec, _ := now()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.rate(sec, int64(nsec)/1e6, 1)
		}
	})
}

// 1700914515 508867500 221728552607000
func TestXXX(t *testing.T) {
	sec, nsec, _ := now()
	r := Rolling{}
	fmt.Println("now()->", r.bucketIdx(sec), r.slotIdx(int64(nsec)/1e6))
	n := time.Now()
	fmt.Println("time.Now()->", r.bucketIdx(n.Unix()), r.slotIdx(n.UnixMilli()))
}
