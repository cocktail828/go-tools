package hystrix

import (
	"context"
	"io"
	"net"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/timex"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	timex.SetTime(func() int64 { return time.Minute.Nanoseconds() })

	assert.NoError(t, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error { return nil },
	))

	v0, v1, _ := GetCircuit(t.Name()).statistic.DualCount(10)
	assert.EqualValues(t, 1, v0)
	assert.EqualValues(t, 0, v1)
}

func TestFailOnError(t *testing.T) {
	assert.Equal(t, net.ErrClosed, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error { return net.ErrClosed },
	))
}

func TestFailOnCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.True(t, slices.Contains([]error{ErrCanceled, net.ErrClosed},
		DoC(
			ctx,
			t.Name(),
			func(ctx context.Context) error { return net.ErrClosed },
		),
	))
}

func TestFailOnTimeout(t *testing.T) {
	GetCircuit(t.Name()).Timeout = 100 * time.Millisecond

	assert.Equal(t, ErrTimeout, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			time.Sleep(time.Minute)
			return nil
		}),
	)
}

func TestMaxConcurrency(t *testing.T) {
	wg := sync.WaitGroup{}
	var good, bad int
	var cnt = 20

	for i := 0; i < cnt; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := DoC(
				context.Background(),
				t.Name(),
				func(ctx context.Context) error {
					time.Sleep(800 * time.Millisecond)
					return nil
				},
			); err == ErrMaxConcurrency {
				bad++
			} else {
				good++
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, cnt-DefaultMaxConcurrency, bad)
	assert.Equal(t, DefaultMaxConcurrency, good)
}

func TestManualy(t *testing.T) {
	t.Run("manual-open", func(t *testing.T) {
		GetCircuit(t.Name()).Trigger(true)
		assert.Equal(t, ErrCircuitOpen, DoC(
			context.Background(),
			t.Name(),
			func(ctx context.Context) error { return nil },
		))
	})

	t.Run("manual-close", func(t *testing.T) {
		GetCircuit(t.Name()).Trigger(false)
		assert.Equal(t, nil, DoC(
			context.Background(),
			t.Name(),
			func(ctx context.Context) error { return nil },
		))
	})
}

// func TestCloseCircuitAfterSuccess(t *testing.T) {
// 	GetCircuit(t.Name().KeepAliveInterval: 50 * time.Millisecond})
// 	cb := GetCircuit(t.Name())
// 	cb.setOpen()

// 	errChan := GoC(context.Background(), t.Name(), fakeFunc(nil))

// 	err := <-errChan
// 	assert.Equal(t, ErrCircuitOpen, err)

// 	for i := 0; i < DefaultRecoveryProbes; i++ {
// 		// wait for allow single test
// 		time.Sleep(50 * time.Millisecond)
// 		<-GoC(context.Background(), t.Name(), fakeFunc(nil))
// 	}
// 	assert.False(t, cb.IsOpen())
// }

func TestFailAfterTimeout(t *testing.T) {
	GetCircuit(t.Name()).Timeout = 10 * time.Millisecond

	assert.Equal(t, ErrTimeout, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			time.Sleep(500 * time.Millisecond)
			return net.ErrClosed
		}),
	)
}

func TestRejected(t *testing.T) {
	GetCircuit(t.Name()).MaxConcurrency = 1
	GetCircuit(t.Name()).tickets.Resize(1)
	GetCircuit(t.Name()).tickets.Acquire(context.TODO(), 1)

	assert.Equal(t, ErrMaxConcurrency, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			// the func will never be called
			panic("the method should never be called")
		}),
	)
}

func TestReturnTicket(t *testing.T) {
	GetCircuit(t.Name()).Timeout = 10 * time.Millisecond

	// the ticket must be returned whether happened
	assert.Equal(t, ErrTimeout, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			select {}
		}),
	)
	assert.Equal(t, 0, GetCircuit(t.Name()).ActiveCount())
}

func TestDoC_Success(t *testing.T) {
	assert.NoError(t, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error { return nil },
	))
}

func TestCircuitOpenOnFail(t *testing.T) {
	GetCircuit(t.Name()).MinQPSThreshold = 2
	assert.NoError(t, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error { return nil },
	))

	for i := 0; i < 3; i++ {
		DoC(
			context.Background(),
			t.Name(),
			func(ctx context.Context) error { return io.ErrClosedPipe },
		)
	}

	GetCircuit(t.Name()).Timeout = 10 * time.Millisecond
	assert.Equal(t, ErrCircuitOpen, DoC(context.Background(), t.Name(), func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}))
}

func TestSingleTest(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })

}
