package rolling

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/cocktail828/go-tools/z/mathx"
)

type counter struct {
	nanots atomic.Int64 // 时间戳，单位为纳秒
	n      atomic.Int64
}

const (
	MIN_COUNTER_NUM  = 128
	MIN_COUNTER_SIZE = 128 // counter 精度, 单位 ms
)

type Rolling struct {
	win      int64 // 计数窗口大小, 单位纳秒, 需为毫秒的2次方幂
	num      int64 // 计数器个数
	bitMask  int64 // num-1
	counters []counter
}

// NewRolling 创建一个新的 DualRolling 实例
// win: 计数器窗口大小，单位：毫秒
// num: 计数器数量，最终会向上取整为 2 的幂次方
func NewRolling(num, win int) *Rolling {
	num = max(int(mathx.Next2Power(int64(num))), MIN_COUNTER_NUM)
	win = max(win, MIN_COUNTER_SIZE)

	return &Rolling{
		win:      int64(win) * 1e6,
		num:      int64(num),
		bitMask:  int64(num - 1),
		counters: make([]counter, num),
	}
}

func (r *Rolling) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Rolling: win:%dns, num:%d\n", r.win, r.num)
	for i := int64(0); i < r.num; i++ {
		current := &r.counters[i]
		if n := current.n.Load(); n > 0 {
			fmt.Fprintf(sb, "%03d => %d\t%vns\n", i, n, current.nanots.Load())
		}
	}
	return sb.String()
}

func (r *Rolling) Incrby(n int) { r._incrby(unixNano(), int64(n)) }
func (r *Rolling) _incrby(nsec, n int64) {
	pos := (nsec / r.win) & r.bitMask
	floor := round(nsec, r.win)
	c := &r.counters[pos]

	for {
		old := c.nanots.Load()
		if old >= floor {
			break
		}

		if c.nanots.CompareAndSwap(old, nsec) {
			c.n.Store(0)
			break
		}
	}
	c.n.Add(n)
}

func (r *Rolling) Count(num int) (int64, int64) { return r._count(unixNano(), int64(num)) }
func (r *Rolling) _count(nsec, num int64) (int64, int64) {
	if num > r.num {
		num = r.num
	}

	var cnt, win int64
	pos := (nsec / r.win) & r.bitMask
	floor := round(nsec, r.win)

	for i := int64(0); i < num; i++ {
		index := (pos - i + r.num) & r.bitMask
		c := &r.counters[index]

		// check whether the counter is expired
		if c.nanots.Load()+r.win*i < floor {
			break
		}

		win++
		cnt += c.n.Load()
	}
	return cnt, win
}

func (r *Rolling) Rate(num int) float64 { return r._rate(unixNano(), int64(num)) }
func (r *Rolling) _rate(nsec, num int64) float64 {
	cnt, win := r._count(nsec, num)
	if win == 0 {
		return 0
	}
	return float64(cnt) / float64(win)
}

func (r *Rolling) QPS(num int) float64 { return r._qps(unixNano(), int64(num)) }
func (r *Rolling) _qps(nsec, num int64) float64 {
	return r._rate(nsec, num) * 1e3 / float64(r.win)
}

func (r *Rolling) Reset() {
	for i := int64(0); i < r.num; i++ {
		r.counters[i].n.Store(0)
		r.counters[i].nanots.Store(0)
	}
}

func (r *Rolling) At(nsec int64) snapshot {
	return snapshot{nsec, r}
}

type snapshot struct {
	nsec int64
	r    *Rolling
}

func (s snapshot) Count(num int) (int64, int64) {
	return s.r._count(s.nsec, int64(num))
}

func (s snapshot) Incrby(n int) {
	s.r._incrby(s.nsec, int64(n))
}

func (s snapshot) QPS(num int) float64 {
	return s.r._qps(s.nsec, int64(num))
}

func (s snapshot) Rate(num int) float64 {
	return s.r._rate(s.nsec, int64(num))
}

func (s snapshot) Reset() {
	s.r.Reset()
}
