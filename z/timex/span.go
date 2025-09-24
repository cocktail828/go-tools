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

// BlockChecker checks we receive at least one msg in d duration. If not, checker
// will print a warn message.
type BlockChecker struct {
	Timeout   time.Duration // required, timeout is the duration to check
	OnTimeout func()        // required, onTimeout is the callback function to call when timeout cannot be nil
	ticker    *time.Ticker
}

// Go starts the check process
// ctx: a cancellable context
func (c *BlockChecker) Go(ctx context.Context) {
	if c.Timeout == 0 || c.OnTimeout == nil {
		panic("timeout or onTimeout is nil")
	}

	go func() {
		c.ticker = time.NewTicker(c.Timeout)
		defer c.ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.ticker.C:
				c.OnTimeout()
			}
		}
	}()
}

// Check resets the time ticker
func (c *BlockChecker) Check() {
	if c.ticker != nil {
		c.ticker.Reset(c.Timeout)
	}
}
