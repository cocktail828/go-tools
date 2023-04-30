package rate_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/rate"
	"github.com/stretchr/testify/assert"
	// "golang.org/x/time/rate"
)

// 代码大部分参考 "golang.org/x/time/rate", 但是修复了并发限不住的bug.
func TestConcurrentRate(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(100)

	var cnt int32 = 0
	limiter := rate.NewLimiter(rate.Every(time.Millisecond), 1)

	closeChan := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-closeChan:
					return
				default:
					if limiter.Allow() {
						atomic.AddInt32(&cnt, 1)
					}
				}
			}
		}()
	}

	time.Sleep(time.Second)
	close(closeChan)
	wg.Wait()
	fmt.Println(cnt) // 大约是 1000
}

func TestRate(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(time.Millisecond*200), 3)
	assert.Equal(t, true, limiter.Allow())
	assert.Equal(t, true, limiter.Allow())
	assert.Equal(t, true, limiter.Allow())
	assert.Equal(t, false, limiter.Allow())

	time.Sleep(time.Second * 2)
	for i := 0; i < 3; i++ {
		assert.Equal(t, true, limiter.Allow())
	}
	time.Sleep(time.Millisecond * 400)
	for i := 0; i < 2; i++ {
		assert.Equal(t, true, limiter.Allow())
	}
	assert.Equal(t, false, limiter.Allow())
}

func TestRateIntf(t *testing.T) {
	limiter := rate.NewLimiter(rate.Inf, 1)
	for i := 0; i < 10000; i++ {
		if !limiter.Allow() {
			panic("oops, should allow")
		}
	}
}

func TestRateNotAllow1(t *testing.T) {
	limiter := rate.NewLimiter(rate.Limit(0), 1)
	limiter.Allow() // consume the bucket
	for i := 0; i < 10000; i++ {
		if limiter.Allow() {
			panic("oops, should not allow")
		}
	}
}

func TestRateNotAllow2(t *testing.T) {
	limiter := rate.NewLimiter(rate.Limit(100), 0)
	for i := 0; i < 10000; i++ {
		if limiter.Allow() {
			panic("oops, should not allow")
		}
	}
}

func BenchmarkRate(b *testing.B) {
	limiter := rate.NewLimiter(rate.Limit(1000000), 1)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			limiter.Allow()
		}
	})
}
