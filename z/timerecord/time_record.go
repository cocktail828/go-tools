package timerecord

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

// TimeOverseer checks we receive at least one msg in d duration. If not, checker
// will print a warn message.
type TimeOverseer struct {
	t      *time.Ticker
	d      time.Duration
	cancel context.CancelFunc
}

// NewTimeOverseer creates a long term checker specified name, checking interval and warning string to print
func NewTimeOverseer(d time.Duration) *TimeOverseer {
	return &TimeOverseer{
		d:      d,
		cancel: func() {},
	}
}

// Start starts the check process
func (c *TimeOverseer) Start() <-chan struct{} {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.t = time.NewTicker(c.d)

	ch := make(chan struct{}, 1)
	go func() {
		defer c.t.Stop()
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case <-c.t.C:
				ch <- struct{}{}
			}
		}
	}()
	return ch
}

// Reset resets the time ticker
func (c *TimeOverseer) Reset() {
	c.t.Reset(c.d)
}

// Stop stops the checker
func (c *TimeOverseer) Stop() {
	c.cancel()
}
