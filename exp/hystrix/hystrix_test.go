package hystrix

import (
	"context"
	"net"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	})

	assert.NoError(t, <-errChan)
	cnt, _ := GetCircuit(t.Name()).statistic.success.Count(1)
	assert.EqualValues(t, 1, cnt)
}

func TestFailure(t *testing.T) {
	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		return net.ErrClosed
	})

	err := <-errChan
	assert.Equal(t, net.ErrClosed, err)
}

func TestCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	errChan := GoC(ctx, t.Name(), func(ctx context.Context) error {
		return net.ErrClosed
	})

	err := <-errChan
	assert.True(t, slices.Contains([]error{ErrCanceled, net.ErrClosed}, err))
}

func TestTimeout(t *testing.T) {
	ConfigureCommand(t.Name(), Setting{Timeout: 100 * time.Millisecond})

	resultChan := make(chan int)
	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		time.Sleep(time.Minute)
		return nil
	})

	assert.Equal(t, 0, len(resultChan))
	assert.Equal(t, ErrTimeout, <-errChan)
}

func TestMaxConcurrency(t *testing.T) {
	ConfigureCommand(t.Name(), Setting{MaxConcurrency: 2})

	run := func(ctx context.Context) error {
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	wg := sync.WaitGroup{}
	var good, bad int
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errChan := GoC(context.Background(), t.Name(), run)
			if err := <-errChan; err == ErrMaxConcurrency {
				bad++
			} else {
				good++
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, 1, bad)
	assert.Equal(t, 2, good)
}

func TestManualy(t *testing.T) {
	GetCircuit(t.Name()).Manually(true)
	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	})
	assert.Equal(t, ErrCircuitOpen, <-errChan)
}

func TestCloseCircuitAfterSuccess(t *testing.T) {
	ConfigureCommand(t.Name(), Setting{MuteWindow: 50 * time.Millisecond})
	cb := GetCircuit(t.Name())
	cb.setOpen()

	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	})

	err := <-errChan
	assert.Equal(t, ErrCircuitOpen, err)

	for i := 0; i < DefaultRecoveryProbes; i++ {
		// wait for allow single test
		time.Sleep(50 * time.Millisecond)
		<-GoC(context.Background(), t.Name(), func(ctx context.Context) error {
			return nil
		})
	}
	assert.False(t, cb.IsOpen())
}

func TestFailAfterTimeout(t *testing.T) {
	ConfigureCommand(t.Name(), Setting{Timeout: 10 * time.Millisecond})

	out := make(chan struct{}, 2)
	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		close(out)
		return net.ErrClosed
	})

	err := <-errChan
	assert.Equal(t, ErrTimeout, err)
	<-out
}

func TestFallbackAfterRejected(t *testing.T) {
	ConfigureCommand(t.Name(), Setting{MaxConcurrency: 1})
	<-GetCircuit(t.Name()).tickets

	runChan := make(chan bool, 1)
	err := <-GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		// the func will never be called
		runChan <- true
		return nil
	})

	assert.Equal(t, ErrMaxConcurrency, err)
	assert.Equal(t, 0, len(runChan))
}

func TestReturnTicket(t *testing.T) {
	ConfigureCommand(t.Name(), Setting{Timeout: 10 * time.Millisecond})

	// the ticket must be returned whether happened
	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		c := make(chan struct{})
		<-c // should block
		return nil
	})

	assert.Equal(t, ErrTimeout, <-errChan)
	assert.Equal(t, 0, GetCircuit(t.Name()).activeCount())
}

func TestDoC_Success(t *testing.T) {
	assert.NoError(t, DoC(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	}))
}

func TestCircuitOpenOnFailure(t *testing.T) {
	ConfigureCommand(t.Name(), Setting{HealthCheckMinReqThreshold: 2})
	assert.NoError(t, DoC(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	}))

	for i := 0; i < 3; i++ {
		DoC(context.Background(), t.Name(), func(ctx context.Context) error {
			return net.ErrClosed
		})
	}

	ConfigureCommand(t.Name(), Setting{Timeout: 10 * time.Millisecond})
	assert.Equal(t, ErrCircuitOpen, DoC(context.Background(), t.Name(), func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}))
}
