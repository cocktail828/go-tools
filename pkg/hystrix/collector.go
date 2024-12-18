package hystrix

import (
	"time"

	"github.com/cocktail828/go-tools/algo/rolling"
)

type Result struct {
	Attempts          int
	Errors            int
	Successes         int
	Failures          int
	Rejects           int
	ShortCircuits     int
	Timeouts          int
	FallbackSuccesses int
	FallbackFailures  int
	CtxCanceled       int
	CtxExceeded       int
	Concurrency       int
	Elapsed           time.Duration
	TotalElapsed      time.Duration
}

// CollectorIMPL holds information about the circuit state.
// This implementation of Collector is the canonical source of information about the circuit.
// It is used for for all internal hystrix operations
// including circuit health checks and metrics sent to the hystrix dashboard.
//
// Metric Collectors do not need Mutexes as they are updated by circuits within a locked context.
type CollectorIMPL struct {
	name                    string
	numRequests             *rolling.Rolling
	errors                  *rolling.Rolling
	successes               *rolling.Rolling
	failures                *rolling.Rolling
	rejects                 *rolling.Rolling
	shortCircuits           *rolling.Rolling
	timeouts                *rolling.Rolling
	contextCanceled         *rolling.Rolling
	contextDeadlineExceeded *rolling.Rolling
	fallbackSuccesses       *rolling.Rolling
	fallbackFailures        *rolling.Rolling
	totalElapsed            *rolling.Timing
	elapsed                 *rolling.Timing
}

func newCollector(name string) *CollectorIMPL {
	m := &CollectorIMPL{name: name}
	m.Reset()
	return m
}

// NumRequests returns the rolling number of requests
func (d *CollectorIMPL) NumRequests() *rolling.Rolling {
	return d.numRequests
}

// Errors returns the rolling number of errors
func (d *CollectorIMPL) Errors() *rolling.Rolling {
	return d.errors
}

// Successes returns the rolling number of successes
func (d *CollectorIMPL) Successes() *rolling.Rolling {
	return d.successes
}

// Failures returns the rolling number of failures
func (d *CollectorIMPL) Failures() *rolling.Rolling {
	return d.failures
}

// Rejects returns the rolling number of rejects
func (d *CollectorIMPL) Rejects() *rolling.Rolling {
	return d.rejects
}

// ShortCircuits returns the rolling number of short circuits
func (d *CollectorIMPL) ShortCircuits() *rolling.Rolling {
	return d.shortCircuits
}

// Timeouts returns the rolling number of timeouts
func (d *CollectorIMPL) Timeouts() *rolling.Rolling {
	return d.timeouts
}

// FallbackSuccesses returns the rolling number of fallback successes
func (d *CollectorIMPL) FallbackSuccesses() *rolling.Rolling {
	return d.fallbackSuccesses
}

func (d *CollectorIMPL) ContextCanceled() *rolling.Rolling {
	return d.contextCanceled
}

func (d *CollectorIMPL) ContextDeadlineExceeded() *rolling.Rolling {
	return d.contextDeadlineExceeded
}

// FallbackFailures returns the rolling number of fallback failures
func (d *CollectorIMPL) FallbackFailures() *rolling.Rolling {
	return d.fallbackFailures
}

// TotalElapsed returns the rolling total duration
func (d *CollectorIMPL) TotalElapsed() *rolling.Timing {
	return d.totalElapsed
}

// Elapsed returns the rolling run duration
func (d *CollectorIMPL) Elapsed() *rolling.Timing {
	return d.elapsed
}

func (d *CollectorIMPL) Update(r Result) {
	d.numRequests.IncrBy(r.Attempts)
	d.errors.IncrBy(r.Errors)
	d.successes.IncrBy(r.Successes)
	d.failures.IncrBy(r.Failures)
	d.rejects.IncrBy(r.Rejects)
	d.shortCircuits.IncrBy(r.ShortCircuits)
	d.timeouts.IncrBy(r.Timeouts)
	d.fallbackSuccesses.IncrBy(r.FallbackSuccesses)
	d.fallbackFailures.IncrBy(r.FallbackFailures)
	d.contextCanceled.IncrBy(r.CtxCanceled)
	d.contextDeadlineExceeded.IncrBy(r.CtxExceeded)
	d.totalElapsed.Add(r.TotalElapsed)
	d.elapsed.Add(r.Elapsed)
}

// Reset resets all metrics in this collector to 0.
func (d *CollectorIMPL) Reset() {
	now := time.Now().Unix()
	d.numRequests.SafeReset(now)
	d.errors.SafeReset(now)
	d.successes.SafeReset(now)
	d.rejects.SafeReset(now)
	d.shortCircuits.SafeReset(now)
	d.failures.SafeReset(now)
	d.timeouts.SafeReset(now)
	d.fallbackSuccesses.SafeReset(now)
	d.fallbackFailures.SafeReset(now)
	d.contextCanceled.SafeReset(now)
	d.contextDeadlineExceeded.SafeReset(now)
	// d.totalElapsed.SafeReset()
	// d.elapsed.SafeReset()
}
