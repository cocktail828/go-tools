// Package rolling implements a high-performance sliding window algorithm for metrics collection
package rolling

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/timex"
	"github.com/stretchr/testify/assert"
)

// TestRollingBasic 测试 Rolling 基本功能
func TestRollingBasic(t *testing.T) {
	timex.SetTime(func() int64 { return time.Minute.Nanoseconds() })
	r := NewRolling(128)
	r.IncrBy(5)

	// 测试计数功能
	cnt, win := r.Count(10)
	assert.EqualValues(t, 5, cnt)
	assert.EqualValues(t, 1, win)

	// 测试多次计数结果一致
	cnt, win = r.Count(10)
	assert.EqualValues(t, 5, cnt)
	assert.EqualValues(t, 1, win)

	// 测试增加更多计数
	r.IncrBy(3)
	cnt, win = r.Count(10)
	assert.EqualValues(t, 8, cnt)
	assert.EqualValues(t, 1, win)
}

// TestDualRollingBasic 测试 DualRolling 基本功能
func TestDualRollingBasic(t *testing.T) {
	timex.SetTime(func() int64 { return time.Minute.Nanoseconds() })
	r := NewRolling(128).Dual()
	r.IncrBy(3, 7) // 增加高32位值3，低32位值7

	// 测试双计数功能
	high, low, win := r.Count(10)
	assert.EqualValues(t, 3, high)
	assert.EqualValues(t, 7, low)
	assert.EqualValues(t, 1, win)

	// 测试多次增加
	r.IncrBy(2, 5)
	high, low, win = r.Count(10)
	assert.EqualValues(t, 5, high)
	assert.EqualValues(t, 12, low)
	assert.EqualValues(t, 1, win)
}

// TestQPS 测试 QPS 计算功能
func TestQPS(t *testing.T) {
	r := NewRolling(128)
	for i := range int64(13) {
		timex.SetTime(func() int64 { return i * _ROLLING_PRECISION })
		r.IncrBy(100 * int(i+1))
	}

	timex.SetTime(func() int64 { return 12 * _ROLLING_PRECISION })
	assert.EqualValues(t, 5859.375, r.QPS(12))

	timex.SetTime(func() int64 { return 0 })
	assert.EqualValues(t, 781.25, r.QPS(1))
	assert.EqualValues(t, 39.0625, r.QPS(20))

	timex.SetTime(func() int64 { return 1000000 * 1e6 })
	assert.EqualValues(t, 0, r.QPS(8))
}

// TestDualRollingQPS 测试 DualRolling 的 QPS 计算功能
func TestDualRollingQPS(t *testing.T) {
	r := NewRolling(128).Dual()
	for i := range int64(13) {
		timex.SetTime(func() int64 { return i * _ROLLING_PRECISION })
		r.IncrBy(10*int(i+1), 20*int(i+1))
	}

	timex.SetTime(func() int64 { return 12 * _ROLLING_PRECISION })
	highQPS, lowQPS := r.QPS(12)
	assert.EqualValues(t, 585.9375, highQPS)
	assert.EqualValues(t, 1171.875, lowQPS)
}

// TestIncrExpire 测试计数器过期功能
func TestIncrExpire(t *testing.T) {
	r := NewRolling(0)
	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(100)

	timex.SetTime(func() int64 { return _ROLLING_MIN_COUNTER * _ROLLING_PRECISION })
	r.IncrBy(23)
	cnt, win := r.Count(1)
	assert.EqualValues(t, 23, cnt)
	assert.EqualValues(t, 1, win)
}

// TestReset 测试重置功能
func TestReset(t *testing.T) {
	r := NewRolling(128)
	dualR := NewRolling(128).Dual()

	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(100)
	dualR.IncrBy(50, 70)

	// 验证重置前的值
	cnt, _ := r.Count(1)
	high, low, _ := dualR.Count(1)
	assert.EqualValues(t, 100, cnt)
	assert.EqualValues(t, 50, high)
	assert.EqualValues(t, 70, low)

	// 执行重置
	r.Reset()
	dualR.Reset()

	// 验证重置后的值
	cnt, _ = r.Count(1)
	high, low, _ = dualR.Count(1)
	assert.EqualValues(t, 0, cnt)
	assert.EqualValues(t, 0, high)
	assert.EqualValues(t, 0, low)
}

// TestSnapshot 测试快照功能
func TestSnapshot(t *testing.T) {
	timex.SetTime(func() int64 { return 100 * _ROLLING_PRECISION })
	r := NewRolling(128)
	r.IncrBy(10)

	// 创建快照
	snapshot := r.At(100 * _ROLLING_PRECISION)

	// 在快照上操作
	snapshot.IncrBy(20)
	cnt, _ := snapshot.Count(1)
	assert.EqualValues(t, 30, cnt)

	// 验证原始计数器也被更新
	cnt, _ = r.Count(1)
	assert.EqualValues(t, 30, cnt)
}

// TestDualSnapshot 测试 DualRolling 快照功能
func TestDualSnapshot(t *testing.T) {
	timex.SetTime(func() int64 { return 100 * _ROLLING_PRECISION })
	r := NewRolling(128).Dual()
	r.IncrBy(5, 10)

	// 创建快照
	snapshot := r.At(100 * _ROLLING_PRECISION)

	// 在快照上操作
	snapshot.IncrBy(15, 20)
	high, low, _ := snapshot.Count(1)
	assert.EqualValues(t, 20, high)
	assert.EqualValues(t, 30, low)

	// 验证原始计数器也被更新
	high, low, _ = r.Count(1)
	assert.EqualValues(t, 20, high)
	assert.EqualValues(t, 30, low)
}

// TestDifferentCounterSizes 测试不同计数器数量的性能和功能
func TestDifferentCounterSizes(t *testing.T) {
	// 测试最小计数器数量
	rMin := NewRolling(1)
	rMin.IncrBy(5)
	cnt, win := rMin.Count(1)
	assert.EqualValues(t, 5, cnt)
	assert.EqualValues(t, 1, win)

	// 测试较大计数器数量
	rLarge := NewRolling(1024)
	rLarge.IncrBy(10)
	cnt, win = rLarge.Count(1)
	assert.EqualValues(t, 10, cnt)
	assert.EqualValues(t, 1, win)
}

// TestBoundaryConditions 测试边界条件
func TestBoundaryConditions(t *testing.T) {
	r := NewRolling(128)

	// 测试零值增加
	r.IncrBy(0)
	cnt, win := r.Count(1)
	assert.EqualValues(t, 0, cnt)
	assert.EqualValues(t, 1, win) // 窗口仍然有效

	// 测试负数增加（如果允许的话）
	r.IncrBy(-5)
	cnt, win = r.Count(1)
	assert.EqualValues(t, -5, cnt)
	assert.EqualValues(t, 1, win)

	// 测试非常大的窗口数量
	cnt, win = r.Count(10000)
	assert.EqualValues(t, -5, cnt)           // 计数不变
	assert.LessOrEqual(t, win, r.numCounter) // 窗口数不超过计数器总数
}

// TestGettime 测试时间设置功能
func TestGettime(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })
	assert.EqualValues(t, 0, timex.UnixNano())

	timex.SetTime(func() int64 { return 1000 })
	assert.EqualValues(t, 1000, timex.UnixNano())
}

// BenchmarkConcurrency 测试并发性能
func BenchmarkConcurrency(b *testing.B) {
	r := NewRolling(0)
	cnt := atomic.Int64{}
	timex.SetTime(func() int64 { return 0 })
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.IncrBy(1)
			cnt.Add(1)
		}
	})

	v, _ := r.Count(1)
	assert.EqualValues(b, v, cnt.Load())
}

// BenchmarkRolling 测试滚动计数器的性能
func BenchmarkRolling(b *testing.B) {
	r := NewRolling(100)

	timex.SetTime(func() int64 { return 0 })
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.QPS(8)
		}
	})
}

// BenchmarkDualRolling 测试双滚动计数器的性能
func BenchmarkDualRolling(b *testing.B) {
	r := NewRolling(100).Dual()

	timex.SetTime(func() int64 { return 0 })
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.QPS(8)
		}
	})
}
