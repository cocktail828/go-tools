package hystrix

import (
	"context"
	"time"

	"github.com/cocktail828/go-tools/pkg/hystrix/event"
	"github.com/pkg/errors"
)

// command models the state used for a single execution on a circuit. "hystrix command" is commonly
// used to describe the pairing of your run/fallback functions with a circuit.
type command struct {
	start   time.Time
	runFn   runFnCtx
	fbFn    fallbackFnCtx
	elapsed time.Duration
	events  []event.Event
	err     error
}

func (c *command) getError() error { return c.err }
func (c *command) report(cb *circuitBreaker) {
	if err := cb.ReportEvent(c.events, c.start, c.elapsed); err != nil {
		log.Printf(err.Error())
	}
}

func (c *command) reportEvent(ev event.Event) { c.events = append(c.events, ev) }

func (c *command) fallback(ctx context.Context, rerr error) {
	c.err = rerr
	if rerr == nil {
		return
	}

	eventType := event.Failure
	switch rerr {
	case ErrCircuitOpen:
		eventType = event.ShortCircuit
	case ErrMaxConcurrency:
		eventType = event.Reject
	case context.Canceled:
		eventType = event.Canceled
	case context.DeadlineExceeded:
		eventType = event.DeadlineExceeded
	}

	c.reportEvent(eventType)
	if c.fbFn == nil {
		return
	}

	c.err = c.fbFn(ctx, rerr)
	if c.err != nil {
		c.reportEvent("fallback-failure")
		c.err = errors.Errorf("fallback failed with '%v'. run error was '%v'", c.err, rerr)
	} else {
		c.reportEvent("fallback-success")
	}
}
