package retry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	err := errors.New("fake error")

	for _, v := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"retry_default_3",
			func(t *testing.T) {
				i := 0
				Do(func() error { i++; return err })
				assert.Equal(t, 3, i)
			},
		},
		{
			"retry_attempt_5",
			func(t *testing.T) {
				i := 0
				Do(func() error { i++; return err }, Attempts(5))
				assert.Equal(t, 5, i)
			},
		},
		{
			"retry_context",
			func(t *testing.T) {
				i := 0
				ctx, cancel := context.WithCancel(context.Background())
				Do(func() error {
					i++
					if i > 2 {
						cancel()
					}
					return err
				}, Context(ctx))
				assert.Equal(t, 3, i)
			},
		},
		{
			"retry_if",
			func(t *testing.T) {
				i := 0
				Do(func() error {
					i++
					return err
				}, Attempts(0), RetryIf(func(attempt uint, err error) bool { return attempt < 3 }))
				assert.Equal(t, 3, i)
			},
		},
	} {
		t.Run(v.name, v.f)
	}
}

// 添加在文件末尾

func TestRetryWithVeryShortTimeout(t *testing.T) {
	// 测试非常短的上下文超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	defer cancel()

	count := 0
	err := Do(func() error {
		count++
		time.Sleep(time.Microsecond * 10) // 确保超时
		return errors.New("fake error")
	}, Attempts(10), Context(ctx))

	assert.Error(t, err)
	assert.True(t, count <= 2) // 应该只尝试1-2次就超时
}

func TestRetryWithOneAttempt(t *testing.T) {
	// 测试重试次数为1的情况
	count := 0
	err := Do(func() error {
		count++
		return errors.New("fake error")
	}, Attempts(1))

	assert.Error(t, err)
	assert.Equal(t, 1, count)
}

func TestRetryWithPanicRecovery(t *testing.T) {
	// 测试在重试过程中函数panic的情况
	count := 0

	// 捕获测试中的panic
	defer func() {
		r := recover()
		assert.Nil(t, r, "测试不应该panic")
	}()

	err := Do(func() error {
		count++
		if count == 2 {
			panic("test panic")
		}
		return errors.New("fake error")
	}, Attempts(3))

	assert.Error(t, err)
	assert.Equal(t, 2, count) // panic应该导致重试停止
}

func TestMultipleOptionsCombination(t *testing.T) {
	err := errors.New("fake error")
	var retryCounts []uint

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	Do(func() error { return err },
		Attempts(4),
		Delay(FixedDelay(time.Millisecond*50)),
		OnRetry(func(attempt uint, err error) {
			retryCounts = append(retryCounts, attempt)
		}),
		RetryIf(func(attempt uint, err error) bool {
			return attempt < 3
		}),
		Context(ctx),
	)

	assert.Len(t, retryCounts, 2) // 应该重试2次
	assert.Equal(t, []uint{1, 2}, retryCounts)
}

func TestErrorWithNilValues(t *testing.T) {
	// 测试包含nil的Error切片
	retryErr := Error{errors.New("error 1"), nil, errors.New("error 3")}

	// 测试Error方法是否正确处理nil错误
	errStr := retryErr.Error()
	assert.Contains(t, errStr, "#1: error 1")
	assert.NotContains(t, errStr, "#2: ") // nil错误不应该显示
	assert.Contains(t, errStr, "#3: error 3")

	// 测试Last方法
	assert.Equal(t, errors.New("error 3"), retryErr.Last())

	// 测试Unwrap方法
	assert.Equal(t, errors.New("error 3"), retryErr.Unwrap())

	// 测试All方法
	allErrs := retryErr.All()
	assert.Len(t, allErrs, 3)
	assert.Equal(t, errors.New("error 1"), allErrs[0])
	assert.Nil(t, allErrs[1])
	assert.Equal(t, errors.New("error 3"), allErrs[2])
}

func TestEmptyError(t *testing.T) {
	// 测试空的Error切片
	retryErr := Error{}

	// 测试Error方法
	assert.Contains(t, retryErr.Error(), "All attempts fail")

	// 测试Last方法
	assert.Nil(t, retryErr.Last())

	// 测试Unwrap方法
	assert.Nil(t, retryErr.Unwrap())

	// 测试All方法
	allErrs := retryErr.All()
	assert.Empty(t, allErrs)

	// 测试Is方法
	assert.False(t, retryErr.Is(errors.New("any error")))

	// 测试As方法
	var targetErr *net.DNSError
	assert.False(t, retryErr.As(&targetErr))
}

func TestDoWithData(t *testing.T) {
	// 测试成功场景
	result, err := DoWithData(func() (string, error) {
		return "success", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "success", result)

	// 测试失败场景
	noperr := errors.New("fake error")
	_, err = DoWithData(func() (string, error) {
		return "", noperr
	}, Attempts(2))
	assert.Error(t, err)
	assert.IsType(t, Error{}, err)

	// 测试重试后成功
	count := 0
	result, err = DoWithData(func() (string, error) {
		count++
		if count < 2 {
			return "", noperr // 使用正确的错误变量
		}
		return "success after retry", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "success after retry", result)
	assert.Equal(t, 2, count)
}

func TestOnRetry(t *testing.T) {
	err := errors.New("fake error")
	var retryCounts []uint
	var retryErrors []error

	OnRetryCallback := func(attempt uint, err error) {
		retryCounts = append(retryCounts, attempt)
		retryErrors = append(retryErrors, err)
	}

	Do(func() error { return err }, Attempts(3), OnRetry(OnRetryCallback))

	assert.Len(t, retryCounts, 2)
	assert.Equal(t, []uint{1, 2}, retryCounts)
	assert.Len(t, retryErrors, 2)
	for _, e := range retryErrors {
		assert.Equal(t, err, e)
	}
}

func TestErrorMethods(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := fmt.Errorf("wrapped: %w", err1)

	// 测试Error接口实现
	retryErr := Error{err1, err2, err3}
	assert.Contains(t, retryErr.Error(), "All attempts fail")
	assert.Contains(t, retryErr.Error(), "#1: error 1")
	assert.Contains(t, retryErr.Error(), "#2: error 2")
	assert.Contains(t, retryErr.Error(), "#3: wrapped: error 1")

	// 测试Last方法
	assert.Equal(t, err3, retryErr.Last())

	// 测试Unwrap方法
	assert.Equal(t, err3, retryErr.Unwrap())

	// 测试Is方法
	assert.True(t, retryErr.Is(err1))
	assert.True(t, retryErr.Is(err2))
	assert.False(t, retryErr.Is(errors.New("not present")))

	// 测试As方法
	var targetErr *net.DNSError // 使用不存在的error类型来测试As方法
	assert.False(t, retryErr.As(&targetErr))

	// 测试All方法
	allErrs := retryErr.All()
	assert.Len(t, allErrs, 3)
	assert.Equal(t, err1, allErrs[0])
	assert.Equal(t, err2, allErrs[1])
	assert.Equal(t, err3, allErrs[2])
}

func TestSuccessOnFirstAttempt(t *testing.T) {
	count := 0
	err := Do(func() error {
		count++
		return nil
	}, Attempts(5))

	assert.NoError(t, err)
	assert.Equal(t, 1, count) // 只应该执行一次
}

func TestRetryWithDelayStrategies(t *testing.T) {
	err := errors.New("fake error")

	// 测试FixedDelay
	startTime := time.Now()
	Do(func() error { return err }, Attempts(2), Delay(FixedDelay(time.Millisecond*50)))
	elapsed := time.Since(startTime)
	assert.GreaterOrEqual(t, elapsed, time.Millisecond*50)

	// 测试RandomDelay - 这里只是简单验证不会崩溃
	Do(func() error { return err }, Attempts(2), Delay(RandomDelay(time.Millisecond*10)))

	// 测试BackoffDelay - 这里只是简单验证不会崩溃
	Do(func() error { return err }, Attempts(2), Delay(BackoffDelay(time.Millisecond*10, time.Millisecond*100)))
}

func TestRetryWithZeroAttempts(t *testing.T) {
	// 当attempts为0时，应该无限重试直到成功或上下文取消
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	count := 0
	err := Do(func() error {
		count++
		return errors.New("fake error")
	}, Attempts(0), Context(ctx))

	assert.Error(t, err)
	assert.Greater(t, count, 0) // 应该尝试了多次
}

func TestConcurrentRetries(t *testing.T) {
	// 简单的并发测试，确保在高并发下不会崩溃
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Do(func() error {
				return errors.New("fake error")
			}, Attempts(2), Delay(FixedDelay(time.Millisecond*10)))
		}()
	}

	wg.Wait()
}

func TestRetryIfWithSpecificError(t *testing.T) {
	specificErr := errors.New("specific error")
	otherErr := errors.New("other error")

	// 只对特定错误重试
	count := 0
	err := Do(func() error {
		count++
		if count == 1 {
			return specificErr
		}
		return otherErr
	}, Attempts(5), RetryIf(func(attempt uint, err error) bool {
		return err == specificErr
	}))

	assert.Equal(t, otherErr, err) // 应该因为otherErr而停止重试
	assert.Equal(t, 2, count)      // 只重试了一次
}

func TestContextCancellation(t *testing.T) {
	// 测试主动取消上下文
	ctx, cancel := context.WithCancel(context.Background())

	count := 0
	done := make(chan struct{})

	go func() {
		Do(func() error {
			count++
			time.Sleep(time.Millisecond * 10) // 给取消操作留出时间
			return errors.New("fake error")
		}, Attempts(10), Context(ctx))
		close(done)
	}()

	// 等待第一次调用后取消上下文
	time.Sleep(time.Millisecond * 15)
	cancel()

	// 等待goroutine完成
	<-done

	assert.Greater(t, count, 0) // 至少尝试了一次
	assert.Less(t, count, 10)   // 没有完成所有10次尝试
}
