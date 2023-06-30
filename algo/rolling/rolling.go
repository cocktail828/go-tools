package rolling

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cocktail828/go-tools/z/fasttime"
)

const (
	_SLOT_SIZE       = 60 // 1 min
	DEFAULT_WIN_SIZE = 3  // 3s
)

type specimen struct {
	rate        int
	inSlotStart int
	inSlotEnd   int
}

type Rolling struct {
	startAt fasttime.Time // 会话创建时间时间
	buckets [2]struct {   // 计数器分奇偶使用, 避免频繁重置
		slots   [_SLOT_SIZE]atomic.Uint32 // 60s
		cleared atomic.Bool               // slot 清零状态标记, 减少重复重置耗时
	}
	busypool *sync.Pool
}

func New() *Rolling {
	return &Rolling{
		startAt: fasttime.Now(),
		busypool: &sync.Pool{
			New: func() interface{} {
				v := make([]specimen, _SLOT_SIZE)
				return &v
			}},
	}
}

func (rolling Rolling) String() string {
	return fmt.Sprintf("Rolling [startAt=%s]", rolling.startAt)
}

func (rolling *Rolling) ResetBucket(t fasttime.TimeIntf, window int) {
	// the bucket may also in use
	if t.Second() <= window {
		return
	}

	// reset the previous bucket
	bucket := &rolling.buckets[(t.Minute()+1)&0x1]
	if bucket.cleared.CompareAndSwap(false, true) {
		for j := 0; j < len(bucket.slots); j++ {
			bucket.slots[j].Store(0)
		}
	}
}

func (rolling *Rolling) Incr() int {
	return rolling.IncrAt(time.Now())
}

func (rolling *Rolling) IncrAt(t fasttime.TimeIntf) int {
	return rolling.IncrByAt(t, 1)
}

func (rolling *Rolling) IncrBy(n int) int {
	return rolling.IncrByAt(time.Now(), n)
}

func (rolling *Rolling) IncrByAt(t fasttime.TimeIntf, n int) int {
	bucket := &rolling.buckets[t.Minute()&0x1]
	bucket.cleared.Store(false)
	return int(bucket.slots[t.Second()].Add(uint32(n)))
}

func (rolling Rolling) get(counterID, slotID, inSlotStart, inSlotEnd int) int {
	if inSlotStart > 0 || inSlotEnd < 1e3 {
		return int(rolling.buckets[counterID].slots[slotID].Load()) * (inSlotEnd - inSlotStart) / 1e3
	}
	return int(rolling.buckets[counterID].slots[slotID].Load())
}

// see https://blog.cloudflare.com/counting-things-a-lot-of-different-things/
// window  in range (0, 5]
func (rolling *Rolling) Rate(t fasttime.Time, window int) float64 {
	alter := t.Add(-time.Second * time.Duration(window))
	if t.UnixMilli()-rolling.startAt.UnixMilli() < int64(window)*1e3 {
		alter = rolling.startAt
	}

	lastIdx := t.Second()
	startMS := alter.MilliSecond()

	pslice := rolling.busypool.Get().(*[]specimen)
	defer rolling.busypool.Put(pslice)
	specimens := (*pslice)[:0]

	for alter.Before(t) {
		s := specimen{
			inSlotStart: startMS,
			inSlotEnd:   1e3,
		}
		startMS = 0

		if lastIdx == alter.Second() {
			s.inSlotEnd = t.MilliSecond()
		}
		s.rate = rolling.get(alter.Minute()&0x1, alter.Second(), s.inSlotStart, s.inSlotEnd)

		specimens = append(specimens, s)
		alter = alter.Add(time.Second)
	}

	sum := 0
	duration := (0) // ms
	for i := 0; i < len(specimens); i++ {
		if specimens[i].rate > 0 {
			sum += specimens[i].rate
			duration += specimens[i].inSlotEnd - specimens[i].inSlotStart
			continue
		}

		// ignore the first zero
		if sum > 0 {
			duration += specimens[i].inSlotEnd - specimens[i].inSlotStart
		}
	}

	if duration == 0 {
		return 0
	}
	return float64(sum * 1e3 / duration)
}
