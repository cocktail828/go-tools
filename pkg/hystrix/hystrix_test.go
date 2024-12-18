package hystrix

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"testing/quick"

	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	resultChan := make(chan struct{})
	errChan := GoC(context.Background(), "", func(ctx context.Context) error {
		close(resultChan)
		return nil
	}, nil)

	<-resultChan
	assert.Equal(t, 0, len(errChan), "no errors should be returned")

	time.Sleep(10 * time.Millisecond)
	assert.EqualValues(t, 1, getCircuit("").statistic.DefaultCollector().Successes().Sum(time.Now()))
}

func TestFallback(t *testing.T) {
	resultChan := make(chan struct{})
	errChan := GoC(context.Background(), "",
		func(ctx context.Context) error { return fmt.Errorf("error") },
		func(ctx context.Context, err error) error {
			if err.Error() == "error" {
				close(resultChan)
			}
			return nil
		})

	<-resultChan
	assert.Equal(t, 0, len(errChan))

	time.Sleep(10 * time.Millisecond)
	cb := getCircuit("")
	assert.EqualValues(t, 0, cb.statistic.DefaultCollector().Successes().Sum(time.Now()))
	assert.EqualValues(t, 1, cb.statistic.DefaultCollector().Failures().Sum(time.Now()))
	assert.EqualValues(t, 1, cb.statistic.DefaultCollector().FallbackSuccesses().Sum(time.Now()))
}

func TestTimeout(t *testing.T) {
	Configure("", Config{Timeout: 100 * time.Millisecond})

	resultChan := make(chan struct{})
	errChan := GoC(context.Background(), "", func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		close(resultChan)
		return nil
	}, func(ctx context.Context, err error) error {
		if err != ErrTimeout && err != context.DeadlineExceeded {
			panic("should be ErrTimeout")
		}
		return nil
	})

	<-resultChan
	assert.Equal(t, 0, len(errChan))
}

func TestTimeoutEmptyFallback(t *testing.T) {
	Configure("", Config{Timeout: 100 * time.Millisecond})

	resultChan := make(chan struct{})
	errChan := GoC(context.Background(), "", func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		close(resultChan)
		return nil
	}, nil)

	assert.Equal(t, ErrTimeout, <-errChan)
	time.Sleep(10 * time.Millisecond)

	cb := getCircuit("")
	assert.EqualValues(t, 0, cb.statistic.DefaultCollector().Successes().Sum(time.Now()))
	assert.EqualValues(t, 1, cb.statistic.DefaultCollector().Timeouts().Sum(time.Now()))
	assert.EqualValues(t, 0, cb.statistic.DefaultCollector().FallbackSuccesses().Sum(time.Now()))
	assert.EqualValues(t, 0, cb.statistic.DefaultCollector().FallbackFailures().Sum(time.Now()))
}

func TestMaxConcurrent(t *testing.T) {
	Configure("", Config{MaxConcurrentRequests: 2})
	resultChan := make(chan struct{})

	run := func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		close(resultChan)
		return nil
	}

	var good, bad int
	for i := 0; i < 3; i++ {
		errChan := GoC(context.Background(), "", run, nil)
		time.Sleep(10 * time.Millisecond)

		select {
		case err := <-errChan:
			if err == ErrMaxConcurrency {
				bad++
			}
		default:
			good++
		}
	}

	assert.Equal(t, 1, bad)
	assert.Equal(t, 2, good)
}

func TestForceOpenCircuit(t *testing.T) {
	cb := getCircuit("")
	cb.Interactive(true)
	errChan := GoC(context.Background(), "", func(ctx context.Context) error {
		return nil
	}, nil)

	assert.Equal(t, ErrCircuitOpen, <-errChan)
	time.Sleep(10 * time.Millisecond)

	assert.EqualValues(t, 0, cb.statistic.DefaultCollector().Successes().Sum(time.Now()))
	assert.EqualValues(t, 1, cb.statistic.DefaultCollector().ShortCircuits().Sum(time.Now()))
}

func TestNilFallbackRunError(t *testing.T) {
	errChan := GoC(context.Background(), "", func(ctx context.Context) error {
		return fmt.Errorf("run_error")
	}, nil)

	err := <-errChan
	assert.Equal(t, err.Error(), "run_error")
}

func TestFailedFallback(t *testing.T) {
	errChan := GoC(context.Background(), "", func(ctx context.Context) error {
		return fmt.Errorf("run_error")
	}, func(ctx context.Context, err error) error {
		return fmt.Errorf("fallback_error")
	})

	err := <-errChan
	assert.Equal(t, "fallback failed with 'fallback_error'. run error was 'run_error'", err.Error())
}

func TestCloseCircuitAfterSuccess(t *testing.T) {
	cb := getCircuit("")
	cb.SetOpen()

	errChan := GoC(context.Background(), "", func(ctx context.Context) error { return nil }, nil)
	assert.Equal(t, ErrCircuitOpen, <-errChan)

	time.Sleep(6 * time.Second)
	done := make(chan bool, 1)
	GoC(context.Background(), "", func(ctx context.Context) error {
		done <- true
		return nil
	}, nil)

	assert.True(t, <-done)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, cb.IsOpen())
}

func TestFailAfterTimeout(t *testing.T) {
	Configure("", Config{Timeout: 10})

	out := make(chan struct{}, 2)
	errChan := GoC(context.Background(), "", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return fmt.Errorf("foo")
	}, func(ctx context.Context, err error) error {
		out <- struct{}{}
		return err
	})

	assert.True(t, strings.Contains((<-errChan).Error(), "timeout"))
	// wait for command to fail, should not panic
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 1, len(out))
}

func TestSlowFallbackOpenCircuit(t *testing.T) {
	Configure("", Config{Timeout: 10})
	cb := getCircuit("")
	cb.SetOpen()

	out := make(chan struct{}, 2)
	GoC(context.Background(), "", func(ctx context.Context) error {
		return nil
	}, func(ctx context.Context, err error) error {
		time.Sleep(100 * time.Millisecond)
		out <- struct{}{}
		return nil
	})

	time.Sleep(250 * time.Millisecond)
	assert.Equal(t, 1, len(out))
	assert.EqualValues(t, 1, cb.statistic.DefaultCollector().ShortCircuits().Sum(time.Now()))
	assert.EqualValues(t, 0, cb.statistic.DefaultCollector().Timeouts().Sum(time.Now()))
}

func TestFallbackAfterRejected(t *testing.T) {
	Configure("", Config{MaxConcurrentRequests: 1})
	cb := getCircuit("")
	cb.TryAcquire()

	runChan := make(chan bool, 1)
	fallbackChan := make(chan bool, 1)
	GoC(context.Background(), "", func(ctx context.Context) error {
		// if run executes after fallback, this will panic due to sending to a closed channel
		runChan <- true
		close(fallbackChan)
		return nil
	}, func(ctx context.Context, err error) error {
		fallbackChan <- true
		close(runChan)
		return nil
	})

	assert.True(t, <-fallbackChan)
	assert.False(t, <-runChan)
}

func TestReturnTicket_QuickCheck(t *testing.T) {
	compareTicket := func() bool {
		Configure("", Config{Timeout: 2})
		errChan := GoC(context.Background(), "", func(ctx context.Context) error {
			c := make(chan struct{})
			<-c // should block
			return nil
		}, nil)
		err := <-errChan
		assert.Equal(t, err, ErrTimeout)
		cb := getCircuit("")
		assert.Nil(t, err)
		return cb.ActiveCount() == 0
	}

	err := quick.Check(compareTicket, nil)
	assert.Nil(t, err)
}

func TestReturnTicket(t *testing.T) {
	Configure("", Config{Timeout: 10})
	errChan := GoC(context.Background(), "", func(ctx context.Context) error {
		c := make(chan struct{})
		<-c // should block
		return nil
	}, nil)

	err := <-errChan
	assert.Equal(t, err, ErrTimeout)

	assert.Nil(t, err)
	assert.Equal(t, getCircuit("").ActiveCount(), 0)
}

func TestContextHandling(t *testing.T) {
	Configure("", Config{Timeout: 15})
	cb := getCircuit("")
	out := make(chan int, 1)
	run := func(ctx context.Context) error {
		time.Sleep(20 * time.Millisecond)
		out <- 1
		return nil
	}

	fallback := func(ctx context.Context, e error) error {
		return nil
	}

	func() {
		errChan := GoC(context.Background(), "", run, nil)
		time.Sleep(25 * time.Millisecond)
		assert.Equal(t, (<-errChan).Error(), ErrTimeout.Error())
		assert.EqualValues(t, cb.statistic.DefaultCollector().NumRequests().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Failures().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Timeouts().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextCanceled().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextDeadlineExceeded().Sum(time.Now()), 0)
	}()

	func() {
		errChan := GoC(context.Background(), "", run, fallback)
		time.Sleep(25 * time.Millisecond)
		assert.Equal(t, len(errChan), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().NumRequests().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Failures().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Timeouts().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextCanceled().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextDeadlineExceeded().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().FallbackSuccesses().Sum(time.Now()), 1)
	}()

	func() {
		testCtx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		errChan := GoC(testCtx, "", run, nil)
		time.Sleep(25 * time.Millisecond)
		assert.Equal(t, (<-errChan).Error(), context.DeadlineExceeded.Error())
		assert.EqualValues(t, cb.statistic.DefaultCollector().NumRequests().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Failures().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Timeouts().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextCanceled().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextDeadlineExceeded().Sum(time.Now()), 1)
		cancel()
	}()

	func() {
		testCtx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		errChan := GoC(testCtx, "", run, fallback)
		time.Sleep(25 * time.Millisecond)
		assert.Equal(t, len(errChan), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().NumRequests().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Failures().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Timeouts().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextCanceled().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextDeadlineExceeded().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().FallbackSuccesses().Sum(time.Now()), 1)
		cancel()
	}()

	func() {
		testCtx, cancel := context.WithCancel(context.Background())
		errChan := GoC(testCtx, "", run, nil)
		time.Sleep(5 * time.Millisecond)
		cancel()
		time.Sleep(20 * time.Millisecond)
		assert.Equal(t, (<-errChan).Error(), context.Canceled.Error())
		assert.EqualValues(t, cb.statistic.DefaultCollector().NumRequests().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Failures().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Timeouts().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextCanceled().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextDeadlineExceeded().Sum(time.Now()), 0)
	}()

	func() {
		testCtx, cancel := context.WithCancel(context.Background())
		errChan := GoC(testCtx, "", run, fallback)
		time.Sleep(5 * time.Millisecond)
		cancel()
		time.Sleep(20 * time.Millisecond)
		assert.Equal(t, len(errChan), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().NumRequests().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Failures().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().Timeouts().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextCanceled().Sum(time.Now()), 1)
		assert.EqualValues(t, cb.statistic.DefaultCollector().ContextDeadlineExceeded().Sum(time.Now()), 0)
		assert.EqualValues(t, cb.statistic.DefaultCollector().FallbackSuccesses().Sum(time.Now()), 1)
	}()
}

func TestDoC(t *testing.T) {
	out := make(chan bool, 1)
	err := DoC(context.Background(), "", func(ctx context.Context) error {
		out <- true
		return nil
	}, nil)
	assert.Nil(t, err)
	assert.True(t, <-out)

	run := func(ctx context.Context) error {
		return fmt.Errorf("i failed")
	}

	err = DoC(context.Background(), "", run, nil)
	assert.Equal(t, "i failed", err.Error())

	err = DoC(context.Background(), "", run, func(ctx context.Context, err error) error {
		out <- true
		return nil
	})
	assert.Nil(t, err)
	assert.True(t, <-out)

	err = DoC(context.Background(), "", run, func(ctx context.Context, err error) error {
		return fmt.Errorf("fallback failed")
	})
	assert.Equal(t, err.Error(), "fallback failed with 'fallback failed'. run error was 'i failed'")

	Configure("", Config{Timeout: 10})
	err = DoC(context.Background(), "", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}, nil)

	assert.Equal(t, err.Error(), "hystrix: timeout")
}
