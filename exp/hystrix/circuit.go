package hystrix

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/cocktail828/go-tools/algo/rolling"
	"github.com/cocktail828/go-tools/pkg/semaphore"
	"github.com/cocktail828/go-tools/z/timex"
)

type keepAlive struct {
	mu    sync.RWMutex
	next  int
	array []bool
}

func (r *keepAlive) Update(ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.array[r.next] = ok
	r.next = (r.next + 1) % len(r.array)
}

func (r *keepAlive) IsHealthy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return !slices.Contains(r.array, false)
}

func (r *keepAlive) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.next = 0
	r.array = make([]bool, len(r.array))
}

type circuitBreaker struct {
	Config
	lastTestAt  atomic.Int64        // last single test timestamp in nanoseconds
	isOpen      atomic.Bool         // circuit state
	isForceOpen bool                // manually turn on/off the circuit
	tickets     *semaphore.Weighted // for concurrency control
	statistic   *rolling.Rolling
	recovery    keepAlive
}

func NewCircuitBreaker(cfg Config) *circuitBreaker {
	cfg.Normalize()
	return &circuitBreaker{
		Config:    cfg,
		tickets:   semaphore.NewWeighted(int64(cfg.MaxConcurrency)),
		statistic: rolling.NewRolling(128),
		recovery: keepAlive{
			array: make([]bool, cfg.KeepAliveProbes),
		},
	}
}

func (cb *circuitBreaker) Update(cfg Config) {
	cb.Config.Update(cfg)
	cb.tickets.Resize(int64(cb.MaxConcurrency))
}

func (cb *circuitBreaker) ActiveCount() int {
	return int(cb.tickets.Assign())
}

// open or close the circuitbreaker manually
func (cb *circuitBreaker) Trigger(toggle bool) {
	cb.isForceOpen = toggle
}

func (cb *circuitBreaker) QPS() float64 {
	return cb.qps(timex.UnixNano())
}

func (cb *circuitBreaker) qps(nsec int64) float64 {
	v0, v1 := cb.statistic.At(nsec).DualQPS(24)
	return v0 + v1
}

func (cb *circuitBreaker) FailRate() int {
	return cb.failRate(timex.UnixNano())
}

func (cb *circuitBreaker) failRate(nsec int64) int {
	var errPct float64
	if cnt0, cnt1, _ := cb.statistic.At(nsec).DualCount(24); cnt0+cnt1 > 0 {
		errPct = float64(cnt1*100) / float64(cnt0+cnt1)
	}
	return int(errPct + 0.5)
}

func (cb *circuitBreaker) allowRequest() bool {
	if cb.isForceOpen {
		return false
	}

	nsec := timex.UnixNano()
	if cb.isOpen.Load() { // recovery?
		if cb.recovery.IsHealthy() {
			cb.setClose()
			return true
		}

		// single test should be applied when circuit is opened
		lastTestAt := cb.lastTestAt.Load()
		if nsec > lastTestAt+cb.KeepAliveInterval.Nanoseconds() {
			swapped := cb.lastTestAt.CompareAndSwap(lastTestAt, nsec)
			if swapped {
				log.Printf("hystrix-go: allowing single test.\n")
			}
			return swapped
		}

		return false
	}

	// regardless of whether it succeeds or not, do not perform a health check when the current qps is too low.
	if int(cb.qps(nsec)) < cb.MinQPSThreshold {
		return true
	}

	// too many failures, open the circuitbreaker
	if cb.failRate(nsec) >= cb.FailureThreshold {
		cb.setOpen()
		return false
	}

	return true
}

func (cb *circuitBreaker) setOpen() {
	if cb.isOpen.CompareAndSwap(false, true) {
		cb.lastTestAt.Store(timex.UnixNano())
		cb.statistic.Reset()
		log.Printf("hystrix-go: opening circuitbreaker.\n")
	}
}

func (cb *circuitBreaker) setClose() {
	if cb.isOpen.CompareAndSwap(true, false) {
		cb.recovery.Reset()
		cb.statistic.Reset()
		log.Printf("hystrix-go: closing circuitbreaker.\n")
	}
}

type summary struct {
	err       error
	startNano int64 // nanoseconds
	stopNano  int64
}

func (cb *circuitBreaker) feedback(s summary) {
	if cb.isOpen.Load() {
		cb.recovery.Update(s.err == nil)
	}

	nsec := s.stopNano
	if s.err == nil {
		cb.statistic.At(nsec).DualIncrBy(1, 0)
	} else {
		cb.statistic.At(nsec).DualIncrBy(0, 1)
	}
}

// GoC runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func (cb *circuitBreaker) GoC(ctx context.Context, runable func(context.Context) error) chan error {
	nanosec := timex.UnixNano()
	errchan := make(chan error, 1)
	if !cb.tickets.TryAcquire(1) {
		cb.feedback(summary{ErrMaxConcurrency, nanosec, nanosec})
		errchan <- ErrMaxConcurrency
		return errchan
	}

	if !cb.allowRequest() {
		cb.feedback(summary{ErrCircuitOpen, nanosec, nanosec})
		errchan <- ErrCircuitOpen
		return errchan
	}

	resultChan := make(chan summary, 1)
	runStart := timex.UnixNano()
	go func() {
		err := runable(ctx)
		resultChan <- summary{err, runStart, timex.UnixNano()}
	}()

	go func() {
		tmoCtx, cancel := context.WithTimeout(context.Background(), cb.Timeout)
		defer cancel()
		defer cb.tickets.Release(1)

		select {
		case rc := <-resultChan:
			errchan <- rc.err
			cb.feedback(rc)

		case <-ctx.Done():
			errchan <- ErrCanceled
			cb.feedback(summary{ErrCanceled, runStart, timex.UnixNano()})

		case <-tmoCtx.Done():
			errchan <- ErrTimeout
			cb.feedback(summary{ErrTimeout, runStart, timex.UnixNano()})
		}
	}()

	return errchan
}
