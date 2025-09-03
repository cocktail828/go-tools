package hystrix

import (
	"context"
	"net"
	"slices"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/timex"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })
	defer timex.ResetTime()

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

func TestFailOnNoTicket(t *testing.T) {
	GetCircuit(t.Name()).MaxConcurrency = 1
	GetCircuit(t.Name()).assigner.Resize(0)

	assert.Equal(t, ErrMaxConcurrency, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			// the func will never be called
			panic("the method should never be called")
		}),
	)
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
	assert.EqualValues(t, 0, GetCircuit(t.Name()).assigner.Allocated())
}

func TestOpenOnTooManyFail(t *testing.T) {
	defer timex.ResetTime()
	timex.SetTime(func() int64 { return 0 })

	GetCircuit(t.Name()).MinQPSThreshold = 2
	timex.SetTime(func() int64 { return 0 })
	for i := 0; i < 100; i++ {
		if i < 80 {
			assert.NoError(t, DoC(
				context.Background(),
				t.Name(),
				func(ctx context.Context) error { return nil },
			))
		} else {
			assert.Equal(t, net.ErrClosed, DoC(
				context.Background(),
				t.Name(),
				func(ctx context.Context) error { return net.ErrClosed },
			))
		}
	}

	assert.Equal(t, 20, GetCircuit(t.Name()).failRate(timex.UnixNano()))
	assert.Equal(t, ErrCircuitOpen, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}),
	)
}

func TestSingleTest(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })
	defer timex.ResetTime()

	GetCircuit(t.Name()).MinQPSThreshold = 2
	timex.SetTime(func() int64 { return 0 })
	for i := 0; i < 100; i++ {
		if i < 80 {
			assert.NoError(t, DoC(
				context.Background(),
				t.Name(),
				func(ctx context.Context) error { return nil },
			))
		} else {
			assert.Equal(t, net.ErrClosed, DoC(
				context.Background(),
				t.Name(),
				func(ctx context.Context) error { return net.ErrClosed },
			))
		}
	}

	assert.Equal(t, 20, GetCircuit(t.Name()).failRate(timex.UnixNano()))
	assert.Equal(t, ErrCircuitOpen, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}),
	)

	for i := 0; i < 5; i++ {
		timex.SetTime(func() int64 { return GetCircuit(t.Name()).KeepAliveInterval.Nanoseconds() * int64(i+1) })
		assert.Equal(t, nil, DoC(
			context.Background(),
			t.Name(),
			func(ctx context.Context) error {
				assert.EqualValues(t, true, ctx.Value(singleTestMeta{}))
				return nil
			}),
		)
	}

	assert.Equal(t, nil, DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			assert.EqualValues(t, false, ctx.Value(singleTestMeta{}))
			return nil
		}),
	)
}
