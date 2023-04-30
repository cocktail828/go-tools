package rolling_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/fasttime"
	"github.com/cocktail828/go-tools/rolling"
)

var (
	r = rolling.New()
)

func TestRolling(t *testing.T) {
	for i := 0; i < 1000; i++ {
		r.Incr()
	}
	time.Sleep(time.Second)
	fmt.Println("result", r.Rate(fasttime.Now(), 1))

	for i := 0; i < 2000; i++ {
		r.Incr()
	}
	time.Sleep(time.Second)
	fmt.Println("result", r.Rate(fasttime.Now(), 1))

	for i := 0; i < 3000; i++ {
		r.Incr()
	}
	time.Sleep(time.Second)
	fmt.Println("result", r.Rate(fasttime.Now(), 1))

	now := fasttime.Now()
	fmt.Printf("%v\n", r)
	fmt.Println("end at:", now.String())
	fmt.Println("result", r.Rate(now, 3))
}

func BenchmarkRolling(b *testing.B) {
	for i := 0; i < 1000; i++ {
		r.Incr()
	}

	time.Sleep(time.Second)
	for i := 0; i < 2000; i++ {
		r.Incr()
	}

	time.Sleep(time.Second)
	for i := 0; i < 3000; i++ {
		r.Incr()
	}

	now := fasttime.Now()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.Rate(now, 3)
		}
	})
}
