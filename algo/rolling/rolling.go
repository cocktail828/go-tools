package rolling

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

const (
	numBucketPerMinute = 60
	leftCleanPos       = numBucketPerMinute >> 2
	rightCleanPos      = (numBucketPerMinute * 3) >> 2
	maxCalcWin         = 3
)

type Rolling struct {
	lastCalcAt time.Time
	buckets    []atomic.Uint32
	cancelFunc context.CancelFunc
	rate       float32
	calcChan   chan time.Time
}

func New() *Rolling {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Rolling{
		lastCalcAt: time.Now(),
		buckets:    make([]atomic.Uint32, numBucketPerMinute),
		cancelFunc: cancel,
		calcChan:   make(chan time.Time, 1),
	}
	go r.startCalc(ctx)
	return r
}

func (r *Rolling) Close() {
	r.cancelFunc()
}

func (r *Rolling) String() string {
	return fmt.Sprintf("Rolling [lastCalcAt=%s]", r.lastCalcAt)
}

func (r *Rolling) reset(now time.Time) {
	sec := now.Second()
	if sec > (leftCleanPos-maxCalcWin) && sec < (leftCleanPos+maxCalcWin) {
		for i := numBucketPerMinute / 2; i < numBucketPerMinute; i++ {
			r.buckets[i].Store(0)
		}
		return
	}

	if sec > (rightCleanPos-maxCalcWin) && sec < (rightCleanPos+maxCalcWin) {
		for i := 0; i < numBucketPerMinute/2; i++ {
			r.buckets[i].Store(0)
		}
	}
}

func (r *Rolling) startCalc(ctx context.Context) {
	timer := time.NewTicker(maxCalcWin)
	for {
		select {
		case <-ctx.Done():
			return

		case now := <-timer.C:
			r.calcChan <- now
			r.reset(now)

		case now := <-r.calcChan:
			r.doCalc(now)
		}
	}
}

// see https://blog.cloudflare.com/counting-things-a-lot-of-different-things/
func (r *Rolling) doCalc(now time.Time) {
	sum := 0
	duration := 0 // ms
	mutable := r.lastCalcAt
	for {
		sec := mutable.Second()
		milliSec := int(mutable.UnixMilli() % 1e3)
		cnt := int(r.buckets[sec].Load())

		if mutable.Before(now) {
			sum += cnt - cnt*milliSec/1e3
			duration += 1e3 - milliSec
		} else {
			sum += cnt * milliSec / 1e3
			duration += milliSec
			break
		}
		mutable = mutable.Add(time.Second)
	}
	r.lastCalcAt = now
	r.rate = float32(sum * 1e3 / duration)
}

func (r *Rolling) Incr() int {
	return r.IncrByAt(time.Now(), 1)
}

func (r *Rolling) IncrAt(t time.Time) int {
	return r.IncrByAt(t, 1)
}

func (r *Rolling) IncrBy(delta int) int {
	return r.IncrByAt(time.Now(), delta)
}

func (r *Rolling) IncrByAt(t time.Time, delta int) int {
	sec := t.Second()
	return int(r.buckets[sec].Add(uint32(delta)))
}

func (r *Rolling) Calc(now time.Time) float32 {
	if now.Sub(r.lastCalcAt) > time.Second {
		select {
		case r.calcChan <- now:
		default:
		}
	}
	return r.rate
}

func (r *Rolling) Rate() float32 {
	return r.rate
}
