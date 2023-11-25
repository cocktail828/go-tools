package rolling

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestCalc(t *testing.T) {
	r := &Rolling{
		calcWnSize:    1,  // 1s
		slotPerBucket: 10, // 100ms
		slots:         make([]atomic.Int64, 640),
		bucketNum:     640,
		bucketSize:    100,
	}
	for i := 0; i < 10; i++ {
		r.slots[i].Add(100)
	}
	for i := 0; i < 10; i++ {
		r.slots[i+10].Add(200)
	}
	for i := 0; i < 10; i++ {
		r.slots[i+20].Add(300)
	}
	fmt.Println(r.calc(640, 110))
	fmt.Println(r.calc(641, 110))
	fmt.Println(r.calc(642, 110))
	r.reset(320, 100)
	fmt.Println(r.slots[:30])
}

func BenchmarkRolling(b *testing.B) {
	r := New(CalcWnSize(1))
	r.IncrBy(1000)
	time.Sleep(time.Second)
	r.IncrBy(2000)
	time.Sleep(time.Second)
	r.IncrBy(3000)
	time.Sleep(time.Second)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			<-r.RateChan()
		}
	})
}

// 1700914515 508867500 221728552607000
// 1700914515 1700914515508 1700914515508875900
func TestXXX(t *testing.T) {
	fmt.Println(now())
	n := time.Now()
	fmt.Println(n.Unix(), n.UnixMilli(), n.UnixNano())
}
