package rolling

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/cocktail828/go-tools/z/mathx"
)

type counter struct {
	tm atomic.Int64 // 时间戳，单位为毫秒
	n  atomic.Uint32
}

const (
	MIN_COUNTER_NUM  = 128
	MIN_COUNTER_SIZE = 128 // counter 精度, 单位 ms
)

type Rolling struct {
	num      int   // Counter 数量，为了快速计算要求为 2 的幂次方
	size     int64 // 计数窗口大小，单位 ms
	bitMask  int   // num-1
	counters []counter
}

func NewRolling(num, size int) *Rolling {
	num = int(mathx.Next2Power(int64(num)))
	if num < MIN_COUNTER_NUM {
		num = MIN_COUNTER_NUM
	}

	if size < MIN_COUNTER_SIZE {
		size = MIN_COUNTER_SIZE
	}

	return &Rolling{
		num:      num,
		size:     int64(size),
		bitMask:  num - 1,
		counters: make([]counter, num),
	}
}

func (r *Rolling) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Rolling: size:%d, num:%d\n", r.size, r.num)
	for i := 0; i < r.num; i++ {
		current := &r.counters[i]
		if n := current.n.Load(); n > 0 {
			fmt.Fprintf(sb, "%03d => %d\t%vms\n", i, n, current.tm.Load())
		}
	}
	return sb.String()
}

func (r *Rolling) Incrby(msec int64, n int) {
	pos := msec / r.size & int64(r.bitMask)
	current := &r.counters[pos]

	for {
		oldTime := current.tm.Load()
		if oldTime+r.size > msec {
			break
		}

		if current.tm.CompareAndSwap(oldTime, msec) {
			current.n.Store(0)
			break
		}
	}

	current.n.Add(uint32(n))
}

func (r *Rolling) Count(msec int64, winsize int) (int64, int64) {
	if winsize > r.num {
		winsize = r.num
	}

	pos := msec / r.size & int64(r.bitMask)
	cnt, win := int64(0), int64(0)

	floor := (msec / r.size) * r.size

	for i := 0; i < winsize; i++ {
		index := (pos - int64(i) + int64(r.num)) & int64(r.bitMask)
		current := &r.counters[index]

		// check whether the counter is expired
		if current.tm.Load()+r.size*int64(i) < floor {
			break
		}

		win++
		cnt += int64(current.n.Load())
	}
	return cnt, win
}

func (r *Rolling) Rate(msec int64, winsize int) float64 {
	cnt, win := r.Count(msec, winsize)
	if win == 0 {
		return 0
	}
	return float64(cnt) / float64(win)
}

func (r *Rolling) QPS(msec int64, winsize int) float64 {
	return r.Rate(msec, winsize) * 1e3 / float64(r.size)
}

func (r *Rolling) Reset() {
	for i := 0; i < r.num; i++ {
		current := &r.counters[i]
		current.n.Store(0)
		current.tm.Store(0)
	}
}
