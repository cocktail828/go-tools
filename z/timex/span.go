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

// Check starts a new event loop check, and resets the timeout to d duration
func NewBlockChecker(d time.Duration, ontmo func(dur time.Duration)) *blockChecker {
	return NewBlockCheckerCtx(context.Background(), d, ontmo)
}

func NewBlockCheckerCtx(ctx context.Context, d time.Duration, ontmo func(dur time.Duration)) *blockChecker {
	subctx, cancel := context.WithCancel(ctx)
	c := &blockChecker{
		cancel:  cancel,
		ticker:  time.NewTicker(d),
		timeout: d,
		ontmo:   ontmo,
	}

	go func() {
		defer c.ticker.Stop()

		timepoint := time.Now()
		for {
			select {
			case <-subctx.Done():
				return
			case <-c.ticker.C:
				c.ontmo(time.Since(timepoint))
			}
		}
	}()

	return c
}

type blockChecker struct {
	cancel  context.CancelFunc
	ticker  *time.Ticker
	timeout time.Duration
	ontmo   func(time.Duration)
}

func (bc *blockChecker) Stop() { bc.cancel() }
func (bc *blockChecker) Ping() { bc.ticker.Reset(bc.timeout) } // I'm still living
