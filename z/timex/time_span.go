package timex

import (
	"context"
	"time"
)

// Recorder provides methods to record time duration
type Recorder struct {
	start time.Time
	last  time.Time
}

// NewTimeRecorder creates a new Recorder
func NewTimeRecorder() *Recorder {
	now := time.Now()
	return &Recorder{now, now}
}

// Duration returns the duration from last record
func (r *Recorder) Duration() time.Duration {
	curr := time.Now()
	span := curr.Sub(r.last)
	r.last = curr
	return span
}

// Elapse returns the duration from the beginning
func (r *Recorder) Elapse() time.Duration {
	curr := time.Now()
	span := curr.Sub(r.start)
	r.last = curr
	return span
}

// LongTermChecker checks we receive at least one msg in d duration. If not, checker
// will print a warn message.
type LongTermChecker struct {
	timeout   time.Duration // timeout is the duration to check
	onTimeout func()        // onTimeout is the callback function to call when timeout cannot be nil
	ticker    *time.Ticker
	ctx       context.Context
	cancel    context.CancelFunc
	elapsed   time.Duration // elapsed is the time duration since Start() is called
}

// NewLongTermChecker creates a long term checker specified name, checking interval and warning string to print
func NewLongTermChecker(timeout time.Duration, onTimeout func()) *LongTermChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &LongTermChecker{
		timeout:   timeout,
		onTimeout: onTimeout,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the check process
func (c *LongTermChecker) Start() {
	go func() {
		c.ticker = time.NewTicker(c.timeout)
		defer c.ticker.Stop()
		startAt := time.Now()
		defer func() { c.elapsed = time.Since(startAt) }()
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-c.ticker.C:
				c.onTimeout()
			}
		}
	}()
}

// Stop stops the checker
func (c *LongTermChecker) Stop() {
	c.cancel()
}

// Check resets the time ticker
func (c *LongTermChecker) Check() {
	if c.ticker != nil {
		c.ticker.Reset(c.timeout)
	}
}

// Elapse report the time duration of the checker
func (c *LongTermChecker) Elapse() time.Duration {
	return c.elapsed
}
