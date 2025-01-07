package rolling

import (
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

// Rolling 滑动窗口计数器
type Rolling struct {
	buckets [BucketNum][SlotPerBucket]atomic.Int64 // 计数器槽
}

// bucketIdx 计算桶索引
func (r *Rolling) bucketIdx(sec int64) int { return int((sec + BucketNum) % BucketNum) }

// slotIdx 计算槽索引
func (r *Rolling) slotIdx(msec int64) int { return int(((msec + 1e3) % 1e3) / 1e2) }

// Recycle 清理过期的桶
// The ticker interval should not exceed `MaxCalcWnSize`
func (r *Rolling) Recycle(c <-chan time.Time) {
	go func() {
		for {
			now, ok := <-c
			if !ok {
				return
			}
			r.SafeReset(now.Unix())
		}
	}()
}

// SafeReset 安全重置过期桶
func (r *Rolling) SafeReset(sec int64) {
	for i := 4; i > 1; i-- {
		bucketIdx := r.bucketIdx(sec - MaxCalcWnSize - int64(i))
		for j := 0; j < SlotPerBucket; j++ {
			r.buckets[bucketIdx][j].Store(0)
		}
	}
}

// sum 计算指定时间窗口内的总和
func (r *Rolling) sum(sec, msec int64, winsize int) int64 {
	if winsize <= 0 || winsize > MaxCalcWnSize {
		winsize = 1 // 默认1秒
	}

	sum := int64(0)
	gap := r.slotIdx(msec)
	for i := 0; i <= winsize; i++ {
		bucketIdx := r.bucketIdx(sec - int64(i))
		for j := 0; j < SlotPerBucket; j++ {
			if (i == winsize && j <= int(gap)) || (i == 0 && j > int(gap)) {
				continue
			}
			sum += r.buckets[bucketIdx][j].Load()
		}
	}

	return sum
}

// SumAt 计算指定时间点的总和
func (r *Rolling) SumAt(now time.Time, winsize int) int64 {
	return r.sum(now.Unix(), now.UnixMilli(), winsize)
}

// Sum 计算当前时间点的总和
func (r *Rolling) Sum(winsize int) int64 {
	sec, nsec, _ := now()
	return r.sum(sec, int64(nsec)/1e6, winsize)
}

// rate 计算指定时间窗口内的速率
func (r *Rolling) rate(sec, msec int64, winsize int) float32 {
	if winsize <= 0 || winsize > MaxCalcWnSize {
		winsize = 1 // 默认1秒
	}
	return float32(r.sum(sec, msec, winsize)) / float32(winsize)
}

// RateAt 计算指定时间点的速率
func (r *Rolling) RateAt(now time.Time, winsize int) float32 {
	return r.rate(now.Unix(), now.UnixMilli(), winsize)
}

// Rate 计算当前时间点的速率
func (r *Rolling) Rate(winsize int) float32 {
	sec, nsec, _ := now()
	return r.rate(sec, int64(nsec)/1e6, winsize)
}

// QPS 计算当前QPS
func (r *Rolling) QPS() float32 {
	sec, nsec, _ := now()
	return r.rate(sec, int64(nsec)/1e6, 1)
}

// Incr 增加计数器
func (r *Rolling) Incr() int64 {
	return r.IncrBy(1)
}

// IncrBy 按指定值增加计数器
func (r *Rolling) IncrBy(delta int) int64 {
	sec, nsec, _ := now()
	bucketIdx := r.bucketIdx(sec)
	slotIdx := r.slotIdx(int64(nsec) / 1e6)
	return r.buckets[bucketIdx][slotIdx].Add(int64(delta))
}
