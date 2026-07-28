package hystrix

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validators(t *testing.T) {
	t.Run("timeout-rejects-negative", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Timeout.Update(-1)
		assert.Equal(t, DefaultTimeout, cfg.Timeout.Val.Get())

		cfg.Timeout.Update(2 * time.Second)
		assert.Equal(t, 2*time.Second, cfg.Timeout.Val.Get())
	})

	t.Run("keepalive-interval-min-1s", func(t *testing.T) {
		cfg := NewConfig()
		cfg.KeepAliveInterval.Update(500 * time.Millisecond) // rejected
		assert.Equal(t, DefaultKeepAliveInterval, cfg.KeepAliveInterval.Val.Get())

		cfg.KeepAliveInterval.Update(3 * time.Second)
		assert.Equal(t, 3*time.Second, cfg.KeepAliveInterval.Val.Get())
	})

	t.Run("keepalive-probes-min-3", func(t *testing.T) {
		cfg := NewConfig()
		cfg.KeepAliveProbes.Update(2) // rejected
		assert.Equal(t, DefaultRecoveryProbes, cfg.KeepAliveProbes.Val.Get())

		cfg.KeepAliveProbes.Update(7)
		assert.Equal(t, 7, cfg.KeepAliveProbes.Val.Get())
	})

	t.Run("max-concurrency-min-1", func(t *testing.T) {
		cfg := NewConfig()
		cfg.MaxConcurrency.Update(0) // rejected
		assert.Equal(t, DefaultMaxConcurrency, cfg.MaxConcurrency.Val.Get())
	})

	t.Run("failure-threshold-range", func(t *testing.T) {
		cfg := NewConfig()
		cfg.FailureThreshold.Update(-1) // rejected
		assert.Equal(t, DefaultFailureThreshold, cfg.FailureThreshold.Val.Get())

		cfg.FailureThreshold.Update(101) // rejected
		assert.Equal(t, DefaultFailureThreshold, cfg.FailureThreshold.Val.Get())

		cfg.FailureThreshold.Update(50) // accepted
		assert.Equal(t, 50, cfg.FailureThreshold.Val.Get())
	})
}

func TestConfig_OnUpdateFires(t *testing.T) {
	cfg := NewConfig()

	var got int
	cfg.MaxConcurrency.OnUpdate.Set(func(v int) { got = v })

	// valid update fires the callback
	cfg.MaxConcurrency.Update(42)
	assert.Equal(t, 42, got)

	// same value: no change, callback must not fire
	got = -1
	cfg.MaxConcurrency.Update(42)
	assert.Equal(t, -1, got)

	// invalid value: rejected before callback
	got = -1
	cfg.MaxConcurrency.Update(0)
	assert.Equal(t, -1, got)
}

func TestConfig_MaxConcurrencyResizesAssigner(t *testing.T) {
	h := NewHystrix(NewConfig())
	assert.EqualValues(t, DefaultMaxConcurrency, h.assigner.Available())

	h.MaxConcurrency.Update(3)
	assert.EqualValues(t, 3, h.assigner.Available())
}

func TestConfig_String(t *testing.T) {
	cfg := NewConfig()
	assert.Equal(t, "10", cfg.MaxConcurrency.String())
}
