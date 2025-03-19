package timex

import (
	"context"
	"time"
)

// TimeRecorder provides methods to record time duration
type TimeRecorder struct {
	start time.Time
	last  time.Time
}

// NewTimeRecorder creates a new TimeRecorder
func NewTimeRecorder() *TimeRecorder {
	return &TimeRecorder{
		start: time.Now(),
		last:  time.Now(),
	}
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
	d       time.Duration
	t       *time.Ticker
	cb      func()
	ctx     context.Context
	cancel  context.CancelFunc
	elapsed time.Duration
}

// NewLongTermChecker creates a long term checker specified name, checking interval and warning string to print
func NewLongTermChecker(d time.Duration, cb func()) *LongTermChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &LongTermChecker{
		d:      d,
		cb:     cb,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the check process
func (c *LongTermChecker) Start() {
	go func() {
		c.t = time.NewTicker(c.d)
		defer c.t.Stop()
		startAt := time.Now()
		defer func() { c.elapsed = time.Since(startAt) }()
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

// Stop stops the checker
func (c *LongTermChecker) Stop() {
	c.cancel()
	if c.t != nil {
		c.t.Stop()
	}
}

// Check resets the time ticker
func (c *LongTermChecker) Check() {
	if c.t != nil {
		c.t.Reset(c.d)
	}
}

// Elapse report the time duration of the checker
func (c *LongTermChecker) Elapse() time.Duration {
	return c.elapsed
}
