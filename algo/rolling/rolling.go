package rolling

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/cocktail828/go-tools/z/mathx"
	"github.com/cocktail828/go-tools/z/timex"
)

const (
	_ROLLING_MIN_COUNTER = 128
	_ROLLING_WINSIZE     = 128
	_ROLLING_PRECISION   = _ROLLING_WINSIZE * 1e6 // 128 ms
)

type Rolling struct {
	numCounter int64 // 计数器个数
	counters   []struct {
		atomic.Int64
		nanots atomic.Int64 // 时间戳，单位为纳秒
	}
}

// NewRolling 创建一个新的滑动计数器实例, 精度为 128 ms
// num: 计数器数量，向上取整为 2 的幂
func NewRolling(num int) *Rolling {
	num = max(int(mathx.Next2Power(int64(num))), _ROLLING_MIN_COUNTER)
	return &Rolling{
		numCounter: int64(num),
		counters: make([]struct {
			atomic.Int64
			nanots atomic.Int64
		}, num),
	}
}

func (r *Rolling) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Rolling: num:%d\n", r.numCounter)
	for i := int64(0); i < r.numCounter; i++ {
		curr := &r.counters[i]
		if n := curr.Load(); n > 0 {
			fmt.Fprintf(sb, "\t%03d => %d\t%vns\n", i, n, curr.nanots.Load())
		}
	}

	return sb.String()
}

func (r *Rolling) indexByTime(nsec int64) int64 {
	return (nsec / _ROLLING_PRECISION) & (r.numCounter - 1)
}

func (r *Rolling) incrBy(nsec, n int64) {
	pos := r.indexByTime(nsec)
	floor := mathx.Floor(nsec, _ROLLING_PRECISION)

	curr := &r.counters[pos]
	for old := curr.nanots.Load(); old < floor; {
		if curr.nanots.CompareAndSwap(old, nsec) {
			curr.Store(0)
			break
		}
	}
	curr.Add(n)
}

func (r *Rolling) count(dual bool, nsec, num int64) (int64, int64, int64) {
	if num > r.numCounter {
		num = r.numCounter
	}

	var cnt0, cnt1, win int64
	edge := r.indexByTime(nsec)
	old := mathx.Floor(nsec, _ROLLING_PRECISION) - _ROLLING_PRECISION*(num-1)

	for i := int64(0); i < num; i++ {
		indexByTime := (edge - i + r.numCounter) & (r.numCounter - 1)
		c := &r.counters[indexByTime]

		// check whether the counter is expired
		if c.nanots.Load() >= old {
			win++

			if dual {
				high, low := mathx.SplitInt64(c.Load())
				cnt0 += int64(high)
				cnt1 += int64(low)
			} else {
				cnt0 += c.Load()
			}
		}
	}
	return cnt0, cnt1, win
}

func (r *Rolling) DualIncrBy(v0, v1 int) {
	r.incrBy(timex.UnixNano(), mathx.MergeInt32(int32(v0), int32(v1)))
}

func (r *Rolling) DualCount(num int) (int64, int64, int64) {
	return r.count(true, timex.UnixNano(), int64(num))
}

func calcQPS(cnt, win int64) float64 {
	if win == 0 {
		return 0
	}
	return float64(cnt) * 1e3 / float64(win) / _ROLLING_WINSIZE
}

func (r *Rolling) DualQPS(num int) (float64, float64) {
	cnt0, cnt1, win := r.count(true, timex.UnixNano(), int64(num))
	return calcQPS(cnt0, win), calcQPS(cnt1, win)
}

func (r *Rolling) IncrBy(n int) {
	r.incrBy(timex.UnixNano(), int64(n))
}

func (r *Rolling) Count(num int) (int64, int64) {
	cnt, _, win := r.count(false, timex.UnixNano(), int64(num))
	return cnt, win
}

func (r *Rolling) QPS(num int) float64 {
	cnt, _, win := r.count(false, timex.UnixNano(), int64(num))
	return calcQPS(cnt, win)
}

func (r *Rolling) At(nsec int64) *snapshot {
	return &snapshot{r, nsec}
}

func (r *Rolling) Reset() {
	for i := range r.counters {
		r.counters[i].Store(0)
	}
}

type snapshot struct {
	ro *Rolling
	tm int64
}

func (r *snapshot) DualIncrBy(v0, v1 int) {
	r.ro.incrBy(r.tm, mathx.MergeInt32(int32(v0), int32(v1)))
}

func (r *snapshot) DualCount(num int) (int64, int64, int64) {
	return r.ro.count(true, r.tm, int64(num))
}

func (r *snapshot) DualQPS(num int) (float64, float64) {
	cnt0, cnt1, win := r.ro.count(true, r.tm, int64(num))
	return calcQPS(cnt0, win), calcQPS(cnt1, win)
}

func (r *snapshot) IncrBy(n int) {
	r.ro.incrBy(r.tm, int64(n))
}

func (r *snapshot) Count(num int) (int64, int64) {
	cnt, _, win := r.ro.count(false, r.tm, int64(num))
	return cnt, win
}

func (r *snapshot) QPS(num int) float64 {
	cnt, _, win := r.ro.count(false, r.tm, int64(num))
	return calcQPS(cnt, win)
}
