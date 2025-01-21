package miscellany

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// TimeRecorder provides methods to record time duration
type TimeRecorder struct {
	start time.Time
	last  time.Time
	meta  sync.Map
}

// NewTimeRecorder creates a new TimeRecorder
func NewTimeRecorder() *TimeRecorder {
	return &TimeRecorder{
		start: time.Now(),
		last:  time.Now(),
	}
}

func (tr *TimeRecorder) Mark(name string) time.Duration {
	span := tr.Duration()
	tr.meta.Store(name, span)
	return span
}

func (tr *TimeRecorder) Load(name string) time.Duration {
	if val, ok := tr.meta.Load(name); ok {
		return val.(time.Duration)
	}
	return 0
}

func (tr *TimeRecorder) Records() map[string]time.Duration {
	meta := map[string]time.Duration{}
	tr.meta.Range(func(key, value any) bool {
		meta[key.(string)] = value.(time.Duration)
		return true
	})
	return meta
}

// Duration returns the duration from last record
func (tr *TimeRecorder) Duration() time.Duration {
	curr := time.Now()
	span := curr.Sub(tr.last)
	tr.last = curr
	return span
}

// Elapse returns the duration from the beginning
func (tr *TimeRecorder) Elapse() time.Duration {
	curr := time.Now()
	span := curr.Sub(tr.start)
	tr.last = curr
	return span
}

// LongTermChecker checks we receive at least one msg in d duration. If not, checker
// will print a warn message.
type LongTermChecker struct {
	d      time.Duration
	t      *time.Ticker
	cb     func() // the `cb` will be called on timeout
	ctx    context.Context
	cancel context.CancelFunc
	start  time.Time
}

// NewLongTermChecker creates a long term checker specified name, checking interval and warning string to print
func NewLongTermChecker(ctx context.Context, d time.Duration, cb func()) *LongTermChecker {
	ctx, cancel := context.WithCancel(ctx)
	return &LongTermChecker{
		d:      d,
		cb:     cb,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the check process
func (c *LongTermChecker) Start() {
	c.t = time.NewTicker(c.d)
	c.start = time.Now()
	go func() {
		defer c.t.Stop()
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-c.t.C:
				c.cb()
			}
		}
	}()
}

// Check resets the time ticker
func (c *LongTermChecker) Check() error {
	if c.t == nil {
		return errors.New("forget call `Start()` ???")
	}
	c.t.Reset(c.d)
	return nil
}

// Stop stops the checker
func (c *LongTermChecker) Stop() {
	c.cancel()
}

// Elapse report the time duration of the checker
func (c *LongTermChecker) Elapse() time.Duration {
	return time.Since(c.start)
}
