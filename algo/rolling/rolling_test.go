package rolling_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/algo/rolling"
)

var (
	r = rolling.New()
)

func TestRolling(t *testing.T) {
	for i := 0; i < 1000; i++ {
		r.Incr()
	}
	time.Sleep(time.Second)
	fmt.Println("result", r.Calc(time.Now()))

	for i := 0; i < 2000; i++ {
		r.Incr()
	}
	time.Sleep(time.Second)
	fmt.Println("result", r.Calc(time.Now()))

	for i := 0; i < 3000; i++ {
		r.Incr()
	}
	time.Sleep(time.Second)
	fmt.Println("result", r.Calc(time.Now()))

	now := time.Now()
	fmt.Printf("%v\n", r)
	fmt.Println("end at:", now.String())
	fmt.Println("result", r.Calc(now))
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

	now := time.Now()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.Calc(now)
		}
	})
}
