package rolling

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/cocktail828/go-tools/z/mathx"
)

type dualcounter struct {
	nanots   atomic.Int64 // 时间戳，单位为纳秒
	negative atomic.Int64
	positive atomic.Int64
}

type DualRolling struct {
	precision  int64 // 计数窗口大小, 单位纳秒, 需为毫秒的2次方幂
	numCounter int64 // 计数器个数
	bitMask    int64 // num-1
	counters   []dualcounter
}

// NewDualRolling 创建一个新的 DualRolling 实例
// win: 计数器窗口大小，单位：毫秒
// num: 计数器数量，最终会向上取整为 2 的幂次方
func NewDualRolling(win, num int) *DualRolling {
	num = int(mathx.Next2Power(int64(num)))
	if num < MIN_COUNTER_NUM {
		num = MIN_COUNTER_NUM
	}

	if win < MIN_COUNTER_SIZE {
		win = MIN_COUNTER_SIZE
	}

	return &DualRolling{
		precision:  int64(win) * 1e6, // ns
		numCounter: int64(num),       // 32.768s
		bitMask:    int64(num) - 1,
		counters:   make([]dualcounter, num),
	}
}

func (r *DualRolling) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Rolling: win:%dns, num:%d\n", r.precision, r.numCounter)
	for i := int64(0); i < r.numCounter; i++ {
		c := &r.counters[i]
		posi := c.positive.Load()
		nega := c.negative.Load()
		if posi > 0 || nega > 0 {
			fmt.Fprintf(sb, "%03d => %d\t%d\t%vns\n", i, posi, nega, c.nanots.Load())
		}
	}
	return sb.String()
}

// Incr 增加成功和失败的计数
func (r *DualRolling) Incr(success, failure int) {
	nsec := unixNano()
	pos := (nsec / r.precision) & r.bitMask
	floor := round(nsec, r.precision)
	c := &r.counters[pos]

	for {
		old := c.nanots.Load()
		if old >= floor {
			break
		}

		if c.nanots.CompareAndSwap(old, nsec) {
			c.positive.Store(0)
			c.negative.Store(0)
			break
		}
	}

	c.positive.Add(int64(success))
	c.negative.Add(int64(failure))
}

// Count 获取过去 num 个窗口的成功和失败次数
func (r *DualRolling) Count(num int64) (posi int, nega int, win int) {
	if num > r.numCounter {
		num = r.numCounter
	}

	nsec := unixNano()
	pos := (nsec / r.precision) & r.bitMask
	limit := round(nsec, r.precision) - r.precision*(num-1)

	for i := int64(0); i < num; i++ {
		index := (pos - i + r.numCounter) & r.bitMask
		c := &r.counters[index]

		// check whether the counter is expired
		if c.nanots.Load() < limit {
			break
		}

		win++
		posi += int(c.positive.Load())
		nega += int(c.negative.Load())
	}
	return
}
