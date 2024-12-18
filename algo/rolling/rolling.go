package rolling

import (
	"context"
	"sync/atomic"
	"time"
	_ "unsafe"
)

const (
	BucketNum     = 64 // 默认64个桶
	SlotPerBucket = 10 // 默认slot个数
	MaxCalcWnSize = 30 // 最大计算窗口(s)
)

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64)

type Rolling struct {
	buckets [BucketNum][SlotPerBucket]atomic.Int64 // counter slots
}

func (r *Rolling) bucketIdx(sec int64) int { return int((sec + BucketNum) % BucketNum) }
func (r *Rolling) slotIdx(msec int64) int  { return int(((msec + 1e3) % 1e3) / 1e2) }

// 清理过期的 bucket
func (r *Rolling) AutoClear(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			r.SafeReset(now.Unix())
		}
	}
}

func (r *Rolling) SafeReset(sec int64) {
	for i := 4; i > 1; i-- {
		bucketIdx := r.bucketIdx(sec - MaxCalcWnSize - int64(i))
		for j := 0; j < SlotPerBucket; j++ {
			r.buckets[bucketIdx][j].Store(0)
		}
	}
}

func (r *Rolling) sum(sec, msec int64, winsize int) int64 {
	if winsize <= 0 || winsize > MaxCalcWnSize {
		winsize = 1 // 1s
	}

	sum := int64(0)
	gap := r.slotIdx(msec)
	for i := winsize; i >= 0; i-- {
		bucketIdx := r.bucketIdx(sec - int64(i))
		for j := 0; j < SlotPerBucket; j++ {
			switch {
			case i == winsize && j <= int(gap):
			case i == 0 && j > int(gap):
			default:
				sum += r.buckets[bucketIdx][j].Load()
			}
		}
	}

	return sum
}

func (r *Rolling) SumAt(now time.Time, winsize int) int64 {
	return r.sum(now.Unix(), now.UnixMilli()/1e6, winsize)
}

func (r *Rolling) Sum(winsize int) int64 {
	sec, nsec, _ := now()
	return r.sum(sec, int64(nsec)/1e6, winsize)
}

// see https://blog.cloudflare.com/counting-things-a-lot-of-different-things/
func (r *Rolling) rate(sec, msec int64, winsize int) float32 {
	if winsize <= 0 || winsize > MaxCalcWnSize {
		winsize = 1 // 1s
	}

	return float32(r.sum(sec, msec, winsize)) / float32(winsize)
}

func (r *Rolling) RateAt(now time.Time, winsize int) float32 {
	return r.rate(now.Unix(), now.UnixMilli(), winsize)
}

func (r *Rolling) Rate(winsize int) float32 {
	sec, nsec, _ := now()
	return r.rate(sec, int64(nsec)/1e6, winsize)
}

func (r *Rolling) QPS() float32 {
	sec, nsec, _ := now()
	return r.rate(sec, int64(nsec)/1e6, 1)
}

func (r *Rolling) Incr() int64 {
	return r.IncrBy(1)
}

func (r *Rolling) IncrBy(delta int) int64 {
	sec, nsec, _ := now()
	bucketIdx := r.bucketIdx(sec)
	slotIdx := r.slotIdx(int64(nsec) / 1e6)
	return r.buckets[bucketIdx][slotIdx].Add(int64(delta))
}
