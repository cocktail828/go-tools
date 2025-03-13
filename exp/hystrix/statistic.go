package hystrix

import (
	"time"

	"github.com/cocktail828/go-tools/algo/rolling"
)

type EventType string

const (
	SuccessEvent        = EventType("success")
	ErrorEvent          = EventType("execute-fail")
	ShortCircuitEvent   = EventType("short-circuit")
	MaxConcurrencyEvent = EventType("max-concurrency")
	TimeoutEvent        = EventType("timeout")
	CancelEvent         = EventType("canceled")
)

type Event struct {
	eventType        EventType
	startAt          time.Time
	stopAt           time.Time
	runDuration      time.Duration
	concurrencyInUse float64
}

type Statistic struct {
	requests *rolling.Rolling
	success  *rolling.Rolling
	failure  *rolling.Rolling
}

func (s *Statistic) Reset() {
	s.requests.Reset()
	s.success.Reset()
	s.failure.Reset()
}

func (s *Statistic) Update(ev Event) {
	msec := ev.stopAt.UnixMilli()
	s.requests.At(msec).Incrby(1)
	if ev.eventType == SuccessEvent {
		s.success.At(msec).Incrby(1)
	} else {
		s.failure.At(msec).Incrby(1)
	}
}

func (s *Statistic) QPS(msec int64) float64 {
	return s.requests.At(msec).QPS(10)
}

func (s *Statistic) FailRate(msec int64) int {
	var errPct float64
	if reqs := s.requests.At(msec).QPS(10); reqs > 0 {
		errs := s.failure.At(msec).QPS(10)
		errPct = errs / reqs * 100
	}

	return int(errPct + 0.5)
}
