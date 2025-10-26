package hystrix

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"slices"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z"
	"github.com/cocktail828/go-tools/z/timex"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })
	defer timex.ResetTime()

	h := NewHystrix(NewConfig())
	assert.NoError(t, h.DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error { return nil },
	))

	v0, v1, _ := h.statistic.Count(10)
	assert.EqualValues(t, 1, v0)
	assert.EqualValues(t, 0, v1)
}

func TestFailOnError(t *testing.T) {
	h := NewHystrix(NewConfig())
	assert.Equal(t, net.ErrClosed, h.DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error { return net.ErrClosed },
	))
}

func TestFailOnCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	h := NewHystrix(NewConfig())
	assert.True(t, slices.Contains([]error{ErrCanceled, net.ErrClosed},
		h.DoC(
			ctx,
			t.Name(),
			func(ctx context.Context) error { return net.ErrClosed },
		),
	))
}

func TestFailOnTimeout(t *testing.T) {
	h := NewHystrix(NewConfig())
	h.Timeout.Update(100 * time.Millisecond)

	assert.Equal(t, ErrTimeout, h.DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			time.Sleep(time.Minute)
			return nil
		}),
	)
}

func TestFailOnNoTicket(t *testing.T) {
	h := NewHystrix(NewConfig())
	h.MaxConcurrency.Update(1)
	h.assigner.Resize(0)

	assert.Equal(t, ErrMaxConcurrency, h.DoC(
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
		h := NewHystrix(NewConfig())
		h.Trigger(true)
		assert.Equal(t, ErrCircuitOpen, h.DoC(
			context.Background(),
			t.Name(),
			func(ctx context.Context) error { return nil },
		))
	})

	t.Run("manual-close", func(t *testing.T) {
		h := NewHystrix(NewConfig())
		h.Trigger(false)
		assert.Equal(t, nil, h.DoC(
			context.Background(),
			t.Name(),
			func(ctx context.Context) error { return nil },
		))
	})
}

func TestReturnTicket(t *testing.T) {
	h := NewHystrix(NewConfig())
	h.Timeout.Update(10 * time.Millisecond)

	// the ticket must be returned whether happened
	assert.Equal(t, ErrTimeout, h.DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			select {}
		}),
	)
	assert.EqualValues(t, 0, h.assigner.Allocated())
}

func TestOpenOnTooManyFail(t *testing.T) {
	defer timex.ResetTime()
	timex.SetTime(func() int64 { return 0 })

	h := NewHystrix(NewConfig())
	h.MinQPSThreshold.Update(2)
	timex.SetTime(func() int64 { return 0 })
	for i := range 100 {
		if i < 80 {
			assert.NoError(t, h.DoC(
				context.Background(),
				t.Name(),
				func(ctx context.Context) error { return nil },
			))
		} else {
			assert.Equal(t, net.ErrClosed, h.DoC(
				context.Background(),
				t.Name(),
				func(ctx context.Context) error { return net.ErrClosed },
			))
		}
	}

	assert.Equal(t, 20, int(h.failRate(timex.UnixNano())))
	assert.Equal(t, ErrCircuitOpen, h.DoC(
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

	h := NewHystrix(NewConfig())
	h.MinQPSThreshold.Update(2)
	timex.SetTime(func() int64 { return 0 })
	for i := range 100 {
		if i < 80 {
			assert.NoError(t, h.DoC(
				context.Background(),
				t.Name(),
				func(ctx context.Context) error { return nil },
			))
		} else {
			assert.Equal(t, net.ErrClosed, h.DoC(
				context.Background(),
				t.Name(),
				func(ctx context.Context) error { return net.ErrClosed },
			))
		}
	}

	assert.Equal(t, 20, int(h.failRate(timex.UnixNano())))
	assert.Equal(t, ErrCircuitOpen, h.DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}),
	)

	for i := range 5 {
		timex.SetTime(func() int64 { return h.KeepAliveInterval.Val.Get().Nanoseconds() * int64(i+1) })
		assert.Equal(t, nil, h.DoC(
			context.Background(),
			t.Name(),
			func(ctx context.Context) error {
				assert.EqualValues(t, true, ctx.Value(singleTestMeta{}), fmt.Sprintf("i: %d", i))
				return nil
			}),
		)
	}

	assert.Equal(t, nil, h.DoC(
		context.Background(),
		t.Name(),
		func(ctx context.Context) error {
			assert.EqualValues(t, false, ctx.Value(singleTestMeta{}))
			return nil
		}),
	)
}

func TestConfigMarshalUnmarshal(t *testing.T) {
	raw := NewConfig()
	raw.Timeout.Update(time.Second * 3)
	raw.KeepAliveInterval.Update(time.Second * 5)
	raw.KeepAliveProbes.Update(10)
	raw.MaxConcurrency.Update(100)
	raw.MinQPSThreshold.Update(10)
	raw.FailureThreshold.Update(50)

	bs, err := json.Marshal(raw)
	z.Must(err)

	cfg := Config{}
	z.Must(json.Unmarshal(bs, &cfg))

	// check if the config is equal
	if !raw.Timeout.Equal(cfg.Timeout) ||
		!raw.KeepAliveInterval.Equal(cfg.KeepAliveInterval) ||
		!raw.KeepAliveProbes.Equal(cfg.KeepAliveProbes) ||
		!raw.MaxConcurrency.Equal(cfg.MaxConcurrency) ||
		!raw.MinQPSThreshold.Equal(cfg.MinQPSThreshold) ||
		!raw.FailureThreshold.Equal(cfg.FailureThreshold) {
		t.Fatalf("unmarshal config failed, expect: %v, got: %v", raw, cfg)
	}
}

func TestConfigUpdate(t *testing.T) {
	raw := NewConfig()
	cfg := NewConfig()
	cfg.FailureThreshold.Update(200) // will not success
	if !raw.FailureThreshold.Equal(cfg.FailureThreshold) {
		t.Fatalf("update FailureThreshold failed, expect: %v, got: %v", raw.FailureThreshold, cfg.FailureThreshold)
	}
}
