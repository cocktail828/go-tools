package rolling

import (
	"context"
	"sync/atomic"
	"time"
	_ "unsafe"
)

const (
	defaultBucketNum     int64 = 64 // 默认64个桶
	defaultSlotPerBucket int64 = 10 // 默认slot个数
	maxCalcWnSize        int64 = 16 // 最大计算窗口(s)
)

type Option func(*Rolling)

// 统计窗口(单位秒)
func CalcWnSize(v int) Option {
	return func(r *Rolling) {
		if 1 <= v && v <= int(maxCalcWnSize) {
			r.calcWnSize = int64(v)
		}
	}
}

// 统计精度, 单位 ms
func Precision(v time.Duration) Option {
	return func(r *Rolling) {
		if v < time.Millisecond*10 {
			r.slotPerBucket = 100
		}
		if v > time.Second {
			r.slotPerBucket = 1
		}
		r.slotPerBucket = int64(v / time.Millisecond / 10)
	}
}

type Rolling struct {
	cancelFunc    context.CancelFunc
	calcWnSize    int64          // second, calc window size
	slotPerBucket int64          // slots per bucket
	slots         []atomic.Int64 // counter slots
	rateCh        chan float32   // rate
	// calced
	bucketNum  int64
	bucketSize int64 // ms
}

func New(opts ...Option) *Rolling {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Rolling{
		calcWnSize:    1, // 1s
		slotPerBucket: defaultSlotPerBucket,
		cancelFunc:    cancel,
		rateCh:        make(chan float32),
	}
	for _, f := range opts {
		f(r)
	}
	r.bucketNum = defaultBucketNum * r.slotPerBucket
	r.bucketSize = 1e3 / r.slotPerBucket
	r.slots = make([]atomic.Int64, r.bucketNum)
	go r.startCalc(ctx)
	return r
}

func (r *Rolling) Close() {
	r.cancelFunc()
}

func (r *Rolling) startCalc(ctx context.Context) {
	timer := time.NewTicker(time.Millisecond * 200)
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-timer.C:
			sec, msec := now.Unix(), now.UnixMilli()%1e3
			r.calc(sec, msec)
			r.reset(sec, msec)
		}
	}
}

func (r *Rolling) reset(sec, msec int64) {
	if sec%maxCalcWnSize == 0 && msec < 500 {
		x := maxCalcWnSize * r.slotPerBucket
		tmp := sec - 2*x
		for i := int64(0); i < x; i++ {
			r.slots[tmp+int64(i)].Store(0)
		}
	}
}

// see https://blog.cloudflare.com/counting-things-a-lot-of-different-things/
func (r *Rolling) calc(sec, msec int64) float32 {
	sum := int64(0)
	duration := float32(r.calcWnSize)

	// fmt.Println(sec, msec)
	start := ((sec-r.calcWnSize)%defaultBucketNum)*r.slotPerBucket + msec/r.bucketSize
	end := (sec%defaultBucketNum)*r.slotPerBucket + msec/r.bucketSize
	if end < start {
		end += r.bucketNum
	}
	zero := true
	for iter := start; iter <= end; iter++ {
		cnt := r.slots[iter%r.bucketNum].Load()
		sum += cnt
		if cnt > 0 {
			zero = false
		}
		if cnt == 0 && zero {
			duration -= float32(1) / float32(r.slotPerBucket)
		}
	}

	// fixup
	sum -= func() int64 {
		x := r.slots[end%r.bucketNum].Load()
		return x - x*(msec%r.bucketSize)/r.bucketSize
	}()
	sum -= r.slots[start].Load() * (msec % r.bucketSize) / r.bucketSize

	rate := float32(0)
	if duration <= 0 {
		rate = 0
	} else {
		rate = float32(sum) / duration
	}
	select {
	case r.rateCh <- rate:
	default:
	}
	return rate
}

func (r *Rolling) Incr() int64 {
	return r.IncrBy(1)
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64)

func (r *Rolling) IncrBy(delta int) int64 {
	sec, nsec, _ := now()
	pos := (sec%defaultBucketNum)*r.slotPerBucket + int64(nsec/1e7)
	return r.slots[pos].Add(int64(delta))
}

func (r *Rolling) RateChan() <-chan float32 {
	return r.rateCh
}
