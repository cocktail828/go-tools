package healthy

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/stretchr/testify/assert"
)

// fakeEvaluater is a controllable Evaluater for exercising keepaliveImpl.
type fakeEvaluater struct {
	alive   atomic.Bool
	checked atomic.Int64
	lastErr atomic.Value // error
}

func (f *fakeEvaluater) Check(err error) {
	f.checked.Add(1)
	if err != nil {
		f.lastErr.Store(err)
	}
}

func (f *fakeEvaluater) Alive() bool { return f.alive.Load() }

// fakeLiveness returns a fixed probe result.
type fakeLiveness struct {
	err    error
	probed atomic.Int64
}

func (f *fakeLiveness) Probe() error {
	f.probed.Add(1)
	return f.err
}

func TestKeepaliveAlive(t *testing.T) {
	ev := &fakeEvaluater{}
	ev.alive.Store(true)
	ka := NewKeepalive(ev, &fakeLiveness{}, xlog.NopPrinter{})

	// First call has a zero cached timestamp, so it must consult the evaluater.
	assert.True(t, ka.Alive())

	// Flip the evaluater to unhealthy. The cached value (<100ms old) is still
	// returned, so Alive stays true until the cache window elapses.
	ev.alive.Store(false)
	assert.True(t, ka.Alive())

	// After the 100ms cache window, Alive must re-evaluate and see the change.
	time.Sleep(120 * time.Millisecond)
	assert.False(t, ka.Alive())
}

func TestKeepaliveCheck(t *testing.T) {
	ev := &fakeEvaluater{}
	ka := NewKeepalive(ev, &fakeLiveness{}, xlog.NopPrinter{})

	errBoom := errors.New("boom")
	ka.Check(errBoom)
	ka.Check(nil)

	assert.Equal(t, int64(2), ev.checked.Load())
	assert.Equal(t, errBoom, ev.lastErr.Load())
}

func TestKeepaliveBackground(t *testing.T) {
	ev := &fakeEvaluater{}
	lv := &fakeLiveness{err: errors.New("probe failed")}
	ka := NewKeepalive(ev, lv, xlog.NopPrinter{})

	cancel := ka.Background(10 * time.Millisecond)
	time.Sleep(120 * time.Millisecond)
	cancel()

	// The background loop should have probed and fed results into the evaluater.
	assert.Greater(t, lv.probed.Load(), int64(0))
	assert.Greater(t, ev.checked.Load(), int64(0))
	assert.Equal(t, lv.err, ev.lastErr.Load())

	// Give the goroutine a moment to observe cancellation and stop.
	probedAfterCancel := lv.probed.Load()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, probedAfterCancel, lv.probed.Load())
}

func TestKeepaliveBackgroundZeroInterval(t *testing.T) {
	ev := &fakeEvaluater{}
	lv := &fakeLiveness{}
	ka := NewKeepalive(ev, lv, xlog.NopPrinter{})

	// A zero interval must not spawn a probing goroutine.
	cancel := ka.Background(0)
	defer cancel()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int64(0), lv.probed.Load())
}
