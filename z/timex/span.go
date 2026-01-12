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

// Reset resets the Recorder to the current time
func (r *Recorder) Reset() {
	r.start = time.Now()
	r.last = r.start
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
}

// Go starts the check process
// The returned Checker can be used to check if the timeout is reached.
// If the ctx is cancelled, the check process will be stopped.
func (c *BlockChecker) Go(ctx context.Context) Checker {
	if c.Timeout <= 0 || c.OnTimeout == nil {
		panic("timeout or onTimeout is nil")
	}

	ticker := time.NewTicker(c.Timeout)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.OnTimeout()
			}
		}
	}()

	return &checkerIMPL{ticker, c.Timeout}
}

type Checker interface {
	// Check resets the time ticker, to avoid ticker timeout
	Check()

	// Reset resets the timeout to d duration
	// It does not reset ot stop the ticker
	Reset(d time.Duration)
}

// checkerIMPL implements Checker interface
type checkerIMPL struct {
	ticker  *time.Ticker
	timeout time.Duration
}

// Check resets the time ticker
func (c *checkerIMPL) Check() {
	c.Reset(c.timeout)
}

// Reset resets the timeout to d duration
func (c *checkerIMPL) Reset(d time.Duration) {
	if d > 0 {
		c.timeout = d
	}
}
