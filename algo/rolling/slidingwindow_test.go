// Package rolling implements a high-performance sliding window algorithm for metrics collection
package rolling

import (
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/timex"
	"github.com/stretchr/testify/assert"
)

// TestRollingBasic 测试 Rolling 基本功能
func TestSlidingWindowBasic(t *testing.T) {
	timex.SetTime(func() int64 { return time.Minute.Nanoseconds() })
	r := NewSlidingWindow(128)
	r.IncrBy(5)

	// 测试计数功能
	cnt, win := r.Estimate(10)
	assert.EqualValues(t, 5, cnt)
	assert.EqualValues(t, 1, win)

	// 测试多次计数结果一致
	cnt, win = r.Estimate(10)
	assert.EqualValues(t, 5, cnt)
	assert.EqualValues(t, 1, win)

	// 测试增加更多计数
	r.IncrBy(3)
	cnt, win = r.Estimate(10)
	assert.EqualValues(t, 8, cnt)
	assert.EqualValues(t, 1, win)
}

// TestQPS 测试 QPS 计算功能
func TestQPS(t *testing.T) {
	r := NewSlidingWindow(128)
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

// TestIncrExpire 测试计数器过期功能
func TestIncrExpire(t *testing.T) {
	r := NewSlidingWindow(0)
	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(100)

	timex.SetTime(func() int64 { return _ROLLING_MIN_COUNTER * _ROLLING_PRECISION })
	r.IncrBy(23)
	cnt, win := r.Estimate(1)
	assert.EqualValues(t, 23, cnt)
	assert.EqualValues(t, 1, win)
}

// TestSnapshot 测试快照功能
func TestSnapshot(t *testing.T) {
	timex.SetTime(func() int64 { return 100 * _ROLLING_PRECISION })
	r := NewSlidingWindow(128)
	r.IncrBy(10)

	// 创建快照
	snapshot := r.At(100 * _ROLLING_PRECISION)

	// 在快照上操作
	snapshot.IncrBy(20)
	cnt, _ := snapshot.Estimate(1)
	assert.EqualValues(t, 30, cnt)

	// 验证原始计数器也被更新
	cnt, _ = r.Estimate(1)
	assert.EqualValues(t, 30, cnt)
}

// TestDifferentCounterSizes 测试不同计数器数量的性能和功能
func TestDifferentCounterSizes(t *testing.T) {
	// 测试最小计数器数量
	rMin := NewSlidingWindow(1)
	rMin.IncrBy(5)
	cnt, win := rMin.Estimate(1)
	assert.EqualValues(t, 5, cnt)
	assert.EqualValues(t, 1, win)

	// 测试较大计数器数量
	rLarge := NewSlidingWindow(1024)
	rLarge.IncrBy(10)
	cnt, win = rLarge.Estimate(1)
	assert.EqualValues(t, 10, cnt)
	assert.EqualValues(t, 1, win)
}

// TestBoundaryConditions 测试边界条件
func TestBoundaryConditions(t *testing.T) {
	r := NewSlidingWindow(128)

	// 测试零值增加
	r.IncrBy(0)
	cnt, win := r.Estimate(1)
	assert.EqualValues(t, 0, cnt)
	assert.EqualValues(t, 1, win) // 窗口仍然有效

	// 测试负数增加（如果允许的话）
	r.IncrBy(-5)
	cnt, win = r.Estimate(1)
	assert.EqualValues(t, -5, cnt)
	assert.EqualValues(t, 1, win)

	// 测试非常大的窗口数量
	cnt, win = r.Estimate(10000)
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
	r := NewSlidingWindow(0)
	cnt := atomic.Int64{}
	timex.SetTime(func() int64 { return 0 })
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.IncrBy(1)
			cnt.Add(1)
		}
	})

	v, _ := r.Estimate(1)
	assert.EqualValues(b, v, cnt.Load())
}

// BenchmarkRolling 测试滚动计数器的性能
func BenchmarkRolling(b *testing.B) {
	r := NewSlidingWindow(100)

	timex.SetTime(func() int64 { return 0 })
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.QPS(8)
		}
	})
}

// TestReset 测试重置功能
func TestReset(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })
	r := NewSlidingWindow(128)
	r.IncrBy(100)

	// 验证重置前的值
	cnt, win := r.Estimate(1)
	assert.EqualValues(t, 100, cnt)
	assert.EqualValues(t, 1, win)

	// 执行重置
	r.Reset()

	// 验证重置后的值（包括时间戳）
	cnt, win = r.Estimate(1)
	assert.EqualValues(t, 0, cnt)
	assert.EqualValues(t, 1, win) // 窗口仍然有效，因为时间戳被重置为0，与当前时间0匹配

	// 重置后可以正常使用
	r.IncrBy(50)
	cnt, win = r.Estimate(1)
	assert.EqualValues(t, 50, cnt)
	assert.EqualValues(t, 1, win)
}

// TestMultipleWindows 测试跨多个窗口的计数
func TestMultipleWindows(t *testing.T) {
	r := NewSlidingWindow(128)

	// 在连续5个窗口中添加计数
	for i := int64(0); i < 5; i++ {
		timex.SetTime(func() int64 { return i * _ROLLING_PRECISION })
		r.IncrBy(int(i+1) * 10)
	}

	// 回到第4个窗口检查最近3个窗口的计数
	timex.SetTime(func() int64 { return 4 * _ROLLING_PRECISION })
	cnt, win := r.Estimate(3)
	assert.EqualValues(t, 3, win)
	// 窗口2: 30, 窗口3: 40, 窗口4: 50 = 120
	assert.EqualValues(t, 120, cnt)

	// 检查所有5个窗口
	cnt, win = r.Estimate(5)
	assert.EqualValues(t, 5, win)
	// 10 + 20 + 30 + 40 + 50 = 150
	assert.EqualValues(t, 150, cnt)
}

// TestTimestampAlignment 测试时间戳对齐到窗口边界
func TestTimestampAlignment(t *testing.T) {
	r := NewSlidingWindow(128)

	// 在同一窗口内的不同时刻添加计数
	baseTime := int64(1000 * _ROLLING_PRECISION)

	// 窗口开始
	timex.SetTime(func() int64 { return baseTime })
	r.IncrBy(10)

	// 窗口中间
	timex.SetTime(func() int64 { return baseTime + _ROLLING_PRECISION/2 })
	r.IncrBy(20)

	// 窗口接近结束
	timex.SetTime(func() int64 { return baseTime + _ROLLING_PRECISION - 1 })
	r.IncrBy(30)

	// 验证所有计数都在同一个窗口
	cnt, win := r.Estimate(1)
	assert.EqualValues(t, 60, cnt)
	assert.EqualValues(t, 1, win)

	// 检查该窗口的时间戳是否对齐到边界
	pos := r.indexByTime(baseTime)
	storedTime := r.counters[pos].nanots.Load()
	expectedFloor := (baseTime / _ROLLING_PRECISION) * _ROLLING_PRECISION
	assert.EqualValues(t, expectedFloor, storedTime, "时间戳应该对齐到窗口边界")
}

// TestWindowExpiration 测试窗口过期机制
func TestWindowExpiration(t *testing.T) {
	r := NewSlidingWindow(128)

	// 在时间0添加计数
	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(100)

	// 前进1个窗口，旧窗口仍然有效
	timex.SetTime(func() int64 { return _ROLLING_PRECISION })
	r.IncrBy(200)
	cnt, win := r.Estimate(2)
	assert.EqualValues(t, 300, cnt)
	assert.EqualValues(t, 2, win)

	// 前进到第128个窗口（刚好绕一圈，旧窗口即将被覆盖）
	timex.SetTime(func() int64 { return 128 * _ROLLING_PRECISION })
	r.IncrBy(50)

	// 旧的第0个窗口应该已经过期
	cnt, win = r.Estimate(200)
	assert.EqualValues(t, 250, cnt) // 只有窗口1和128的计数
	assert.EqualValues(t, 2, win)
}

// TestConcurrentWindowReset 测试并发窗口重置的正确性
func TestConcurrentWindowReset(t *testing.T) {
	r := NewSlidingWindow(128)
	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(100)

	// 切换到新窗口
	timex.SetTime(func() int64 { return _ROLLING_PRECISION })

	// 并发地在新窗口添加计数
	const goroutines = 10
	const incrementsPerGoroutine = 100
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < incrementsPerGoroutine; j++ {
				r.IncrBy(1)
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// 验证新窗口的计数正确
	cnt, _ := r.Estimate(1)
	expected := int64(goroutines * incrementsPerGoroutine)
	assert.EqualValues(t, expected, cnt, "并发增加应该得到正确的总数")

	// 验证旧窗口的计数仍然存在
	cnt, win := r.Estimate(2)
	assert.EqualValues(t, 100+expected, cnt)
	assert.EqualValues(t, 2, win)
}

// TestSnapshotIsolation 测试快照的独立性
func TestSnapshotIsolation(t *testing.T) {
	r := NewSlidingWindow(128)

	// 在时间100添加计数
	timex.SetTime(func() int64 { return 100 * _ROLLING_PRECISION })
	r.IncrBy(10)

	// 创建在时间100的快照
	snap100 := r.At(100 * _ROLLING_PRECISION)

	// 时间前进到101（相邻窗口）
	timex.SetTime(func() int64 { return 101 * _ROLLING_PRECISION })
	r.IncrBy(20)

	// 创建在时间101的快照
	snap101 := r.At(101 * _ROLLING_PRECISION)

	// 快照100应该只看到时间100的数据
	cnt, win := snap100.Estimate(1)
	assert.EqualValues(t, 10, cnt)
	assert.EqualValues(t, 1, win)

	// 快照101应该看到两个窗口的数据
	cnt, win = snap101.Estimate(2)
	assert.EqualValues(t, 30, cnt)
	assert.EqualValues(t, 2, win)

	// 在快照100上添加计数，应该影响时间100的窗口
	snap100.IncrBy(5)
	cnt, _ = snap100.Estimate(1)
	assert.EqualValues(t, 15, cnt)

	// 但这也应该反映在当前计数中
	timex.SetTime(func() int64 { return 101 * _ROLLING_PRECISION })
	cnt, win = r.Estimate(2)
	assert.EqualValues(t, 35, cnt) // 15 + 20
	assert.EqualValues(t, 2, win)
}

// TestQPSCalculation 测试 QPS 计算的准确性
func TestQPSCalculation(t *testing.T) {
	r := NewSlidingWindow(128)

	// 在1秒内（约8个窗口）添加1000个请求
	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(1000)

	// QPS = 1000 * 1000 / (1 * 128) = 7812.5
	qps := r.QPS(1)
	assert.InDelta(t, 7812.5, qps, 0.01)

	// 在下一个窗口添加500个请求
	timex.SetTime(func() int64 { return _ROLLING_PRECISION })
	r.IncrBy(500)

	// 2个窗口的 QPS = 1500 * 1000 / (2 * 128) = 5859.375
	qps = r.QPS(2)
	assert.InDelta(t, 5859.375, qps, 0.01)
}

// TestCountWithInvalidWindows 测试包含无效窗口的计数
func TestCountWithInvalidWindows(t *testing.T) {
	r := NewSlidingWindow(128)

	// 只在几个窗口添加数据
	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(10)

	timex.SetTime(func() int64 { return 2 * _ROLLING_PRECISION })
	r.IncrBy(20)

	timex.SetTime(func() int64 { return 5 * _ROLLING_PRECISION })
	r.IncrBy(30)

	// 请求10个窗口，从窗口5往回数
	// 环形缓冲区会检查10个连续的窗口位置
	// 窗口5,4,3,2,1,0,-1,-2,-3,-4 (对应索引 5,4,3,2,1,0,127,126,125,124)
	// 但是只有窗口 5,2,0 有非零时间戳，所以实际计入的是这3个窗口
	// 注意：由于 Reset 后时间戳为0，窗口位置1,3,4 的时间戳为0，也 >= old（负数）
	// 所以实际上会计入所有时间戳 >= old 的窗口
	cnt, win := r.Estimate(10)
	assert.EqualValues(t, 60, cnt)
	// 窗口5,4,3,2,1,0的时间戳都满足 >= old 的条件（窗口-1,-2,-3,-4的时间戳也是0，也满足）
	// 实际上所有10个位置的时间戳都是0或正数，都 >= old（负数）
	assert.EqualValues(t, 10, win) // 所有10个窗口位置都被认为"有效"

	// QPS 应该基于实际有效窗口计算
	qps := r.QPS(10)
	expectedQPS := float64(60) * 1000 / float64(10) / 128
	assert.InDelta(t, expectedQPS, qps, 0.01)
}

// TestRingBufferWrap 测试环形缓冲区的循环覆盖
func TestRingBufferWrap(t *testing.T) {
	r := NewSlidingWindow(128)

	// 填满整个环形缓冲区
	for i := int64(0); i < 128; i++ {
		timex.SetTime(func() int64 { return i * _ROLLING_PRECISION })
		r.IncrBy(1)
	}

	cnt, win := r.Estimate(128)
	assert.EqualValues(t, 128, cnt)
	assert.EqualValues(t, 128, win)

	// 再前进一个窗口，应该覆盖最早的窗口
	timex.SetTime(func() int64 { return 128 * _ROLLING_PRECISION })
	r.IncrBy(10)

	// 请求129个窗口，但最多只有128个
	cnt, win = r.Estimate(129)
	assert.EqualValues(t, 137, cnt) // 127个窗口(1-127) + 1个新窗口(128)的10
	assert.EqualValues(t, 128, win)

	// 验证窗口0被正确覆盖
	pos := r.indexByTime(128 * _ROLLING_PRECISION)
	assert.EqualValues(t, 0, pos) // 应该回到位置0
	assert.EqualValues(t, 10, r.counters[pos].Load())
}

// TestCountExceedCapacity 测试请求窗口数超过缓冲区容量时，
// 超出部分一律计 0，且有效窗口数不超过实际写入的窗口。
func TestCountExceedCapacity(t *testing.T) {
	r := NewSlidingWindow(128) // 容量 128

	// 基准放在远离 0 的窗口，避免空槽(nanots=0)落入有效区间的 ts=0 歧义。
	const base = 1000
	for i := int64(0); i < 3; i++ {
		timex.SetTime(func() int64 { return (base + i) * _ROLLING_PRECISION })
		r.IncrBy(10)
	}

	timex.SetTime(func() int64 { return (base + 2) * _ROLLING_PRECISION })

	// 请求远超容量(1000 >> 128)，超出容量的窗口物理上不存在，应计 0。
	// 实际有效窗口只有写入过的 3 个。
	cnt, win := r.Estimate(1000)
	assert.EqualValues(t, 30, cnt)
	assert.EqualValues(t, 3, win)

	// 恰好等于容量，结果一致。
	cnt, win = r.Estimate(128)
	assert.EqualValues(t, 30, cnt)
	assert.EqualValues(t, 3, win)

	// 超过容量时不会因环形绕回而重复计数：win 绝不超过 numCounter。
	_, win = r.Estimate(math.MaxInt32)
	assert.LessOrEqual(t, win, r.numCounter)
}

// TestCountNonPositive 测试非正数窗口请求返回 (0,0) 而不崩溃。
func TestCountNonPositive(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })
	r := NewSlidingWindow(128)
	r.IncrBy(5)

	cnt, win := r.Estimate(0)
	assert.EqualValues(t, 0, cnt)
	assert.EqualValues(t, 0, win)

	cnt, win = r.Estimate(-100)
	assert.EqualValues(t, 0, cnt)
	assert.EqualValues(t, 0, win)

	assert.EqualValues(t, 0, r.QPS(0))
	assert.EqualValues(t, 0, r.QPS(-1))
}

// TestNewSlidingWindowNonPositive 测试非正数容量被兜底到最小容量而不 panic。
func TestNewSlidingWindowNonPositive(t *testing.T) {
	for _, n := range []int{-1, 0, -1024, math.MinInt32} {
		assert.NotPanics(t, func() {
			r := NewSlidingWindow(n)
			assert.EqualValues(t, _ROLLING_MIN_COUNTER, r.numCounter)
		}, "NewSlidingWindow(%d) 不应 panic", n)
	}
}

// TestTimeGoesBackward 测试时间倒退时增量被安全丢弃，不污染已占用的窗口。
// 窗口 128 与窗口 0 映射到同一个 slot(0)，先占用窗口 128，再倒退到窗口 0，
// 触发 old > floor 分支。
func TestTimeGoesBackward(t *testing.T) {
	r := NewSlidingWindow(128)

	// 先在窗口 128 写入，slot 0 的时间戳被设为 128*P
	timex.SetTime(func() int64 { return 128 * _ROLLING_PRECISION })
	r.IncrBy(100)
	assert.EqualValues(t, 0, r.indexByTime(128*_ROLLING_PRECISION))

	// 时间倒退到窗口 0(同样映射到 slot 0，floor=0 < old=128*P)
	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(999) // 应被丢弃，不能覆盖 slot 0 的现有窗口

	// slot 0 仍属于窗口 128，值不变
	assert.EqualValues(t, 100, r.counters[0].Load())
	assert.EqualValues(t, 128*_ROLLING_PRECISION, r.counters[0].nanots.Load())

	// 窗口 128 的数据完好
	cnt, win := r.At(128 * _ROLLING_PRECISION).Estimate(1)
	assert.EqualValues(t, 100, cnt)
	assert.EqualValues(t, 1, win)
}

// TestRolloverConcurrentNoLoss 回归测试：窗口翻转(slot 复用)时并发写入不丢更新。
// 复现的 bug 是抢窗口的 goroutine 用 Store 覆盖了其他 goroutine 的 Add。
// 用 -race 运行效果更佳。
func TestRolloverConcurrentNoLoss(t *testing.T) {
	const iterations = 200
	const goroutines = 500

	for iter := 0; iter < iterations; iter++ {
		r := NewSlidingWindow(128)

		// 在窗口 0(slot 0)写入陈旧值
		timex.SetTime(func() int64 { return 0 })
		r.IncrBy(1_000_000)

		// 跳到窗口 128(同样是 slot 0)，并发写入 goroutines 次 IncrBy(1)
		timex.SetTime(func() int64 { return 128 * _ROLLING_PRECISION })
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() { defer wg.Done(); r.IncrBy(1) }()
		}
		wg.Wait()

		cnt, _ := r.At(128 * _ROLLING_PRECISION).Estimate(1)
		if cnt != goroutines {
			t.Fatalf("iter %d: 期望 %d, 实际 %d (丢失 %d)",
				iter, goroutines, cnt, int64(goroutines)-cnt)
		}
	}
}

// TestString 测试 String 输出包含非零窗口信息。
func TestString(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })
	r := NewSlidingWindow(128)

	// 空窗口只输出头部
	assert.Contains(t, r.String(), "SlidingWindow:")

	r.IncrBy(42)
	s := r.String()
	assert.Contains(t, s, "SlidingWindow:")
	assert.Contains(t, s, "42") // 非零计数应出现
}

// TestSnapshotQPS 测试快照的 QPS 计算(含 win==0 分支)。
func TestSnapshotQPS(t *testing.T) {
	r := NewSlidingWindow(128)

	timex.SetTime(func() int64 { return 0 })
	r.IncrBy(1000)

	// 快照在有数据的时刻
	qps := r.At(0).QPS(1)
	assert.InDelta(t, 7812.5, qps, 0.01) // 1000*1000/(1*128)

	// 快照在无数据的遥远时刻，win==0，QPS 应为 0
	assert.EqualValues(t, 0, r.At(10000*_ROLLING_PRECISION).QPS(1))
}
