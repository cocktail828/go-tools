package hystrix

import (
	"sync"
	"time"

	"github.com/cocktail828/go-tools/algo/rolling"
	"github.com/cocktail828/go-tools/pkg/hystrix/event"
	"github.com/cocktail828/go-tools/z/locker"
)

type temporary struct {
	Types       []event.Event `json:"types,omitempty"`
	Start       time.Time     `json:"start,omitempty"`
	Elapsed     time.Duration `json:"elapsed,omitempty"`
	Concurrency int           `json:"concurrency,omitempty"`
}

type statistic struct {
	sync.RWMutex
	Name      string
	Chan      chan temporary
	Collector *CollectorIMPL
}

func newStatistic(name string) *statistic {
	m := &statistic{
		Name:      name,
		Chan:      make(chan temporary, 2000),
		Collector: newCollector(name),
	}
	go m.Monitor()
	return m
}

func (m *statistic) Monitor() {
	for s := range m.Chan {
		locker.WithRLock(m, func() {
			totalDuration := time.Since(s.Start)
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go m.IncrementMetrics(wg, m.Collector, s, totalDuration)
			wg.Wait()
		})
	}
}

func (m *statistic) IncrementMetrics(wg *sync.WaitGroup, collector Collector, s temporary, totalDuration time.Duration) {
	r := Result{
		Attempts:     1,
		Elapsed:      s.Elapsed,
		Concurrency:  s.Concurrency,
		TotalElapsed: totalDuration,
	}

	switch s.Types[0] {
	case "success":
		r.Successes = 1
	case "failure":
		r.Failures = 1
		r.Errors = 1
	case "rejected":
		r.Rejects = 1
		r.Errors = 1
	case "short-circuit":
		r.ShortCircuits = 1
		r.Errors = 1
	case "timeout":
		r.Timeouts = 1
		r.Errors = 1
	case "context_canceled":
		r.CtxCanceled = 1
	case "context_deadline_exceeded":
		r.CtxExceeded = 1
	}

	// fallback metrics
	if len(s.Types) > 1 {
		if s.Types[1] == "fallback-success" {
			r.FallbackSuccesses = 1
		}
		if s.Types[1] == "fallback-failure" {
			r.FallbackFailures = 1
		}
	}

	collector.Update(r)
	wg.Done()
}

func (m *statistic) requestsLocked() *rolling.Rolling {
	return m.Collector.NumRequests()
}

func (m *statistic) Requests() (n *rolling.Rolling) {
	locker.WithLock(m, func() {
		n = m.requestsLocked()
	})
	return
}

func (m *statistic) ErrorPercent(now time.Time) (n int) {
	locker.WithLock(m, func() {
		reqs := m.requestsLocked().SumAt(now, 1)
		errs := m.Collector.Errors().SumAt(now, 1)

		var errPct float64
		if reqs > 0 {
			errPct = (float64(errs) / float64(reqs)) * 100
		}
		n = int(errPct + 0.5)
	})
	return
}

func (m *statistic) IsHealthy(now time.Time) bool {
	return m.ErrorPercent(now) < getCircuit(m.Name).ErrorPercentThreshold
}
