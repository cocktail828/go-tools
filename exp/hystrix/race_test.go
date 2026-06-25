package hystrix

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestRaceCondition_IsForceOpen 测试 isForceOpen 字段的数据竞态
func TestRaceCondition_IsForceOpen(t *testing.T) {
	h := NewHystrix(NewConfig())

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: 不断修改 isForceOpen
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			h.Trigger(i%2 == 0)
			time.Sleep(time.Microsecond)
		}
	}()

	// Goroutine 2: 不断读取 isForceOpen (通过 allowRequest)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			h.allowRequest()
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Wait()
}

// TestGoroutineLeak_ResultChan 测试 resultChan goroutine 泄漏
func TestGoroutineLeak_ResultChan(t *testing.T) {
	cfg := NewConfig()
	cfg.Timeout.Val.Set(50 * time.Millisecond)
	h := NewHystrix(cfg)

	initialGoroutines := countGoroutines()
	t.Logf("Initial goroutines: %d", initialGoroutines)

	// 创建 100 个会超时的请求
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		errChan := h.GoC(ctx, func(ctx context.Context) error {
			// 模拟慢操作，会超时
			time.Sleep(200 * time.Millisecond)
			return nil
		})

		<-errChan // 等待错误（应该是超时）
		cancel()
	}

	// 等待一段时间让 goroutine 有机会退出
	time.Sleep(500 * time.Millisecond)

	finalGoroutines := countGoroutines()
	t.Logf("Final goroutines: %d", finalGoroutines)

	leaked := finalGoroutines - initialGoroutines
	if leaked > 10 {
		t.Errorf("Potential goroutine leak: %d goroutines leaked", leaked)
	}
}

// TestConcurrentTrigger 测试并发调用 Trigger
func TestConcurrentTrigger(t *testing.T) {
	h := NewHystrix(NewConfig())

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				h.Trigger(j%2 == 0)
				h.allowRequest()
			}
		}(i)
	}

	wg.Wait()
}

// countGoroutines 统计当前 goroutine 数量（简单实现）
func countGoroutines() int {
	return runtime.NumGoroutine()
}
