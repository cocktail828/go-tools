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

const (
	_COUNTER_WIN_SIZE = 24
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

type Metric struct {
	Concurrency int     // 并发数
	FailRate    int     // 0~100
	QPS         float64 // 5s 内的 QPS
}

func (cb *circuitBreaker) Metric() Metric {
	return Metric{
		Concurrency: int(cb.tickets.Assign()),
		FailRate:    cb.failRate(timex.UnixNano()),
		QPS:         cb.qps(timex.UnixNano()),
	}
}

// open or close the circuitbreaker manually
func (cb *circuitBreaker) Trigger(isopen bool) {
	cb.isForceOpen = isopen
}

func (cb *circuitBreaker) qps(nsec int64) float64 {
	v0, v1 := cb.statistic.At(nsec).DualQPS(_COUNTER_WIN_SIZE)
	return v0 + v1
}

func (cb *circuitBreaker) failRate(nsec int64) int {
	var errPct float64
	if cnt0, cnt1, _ := cb.statistic.At(nsec).DualCount(_COUNTER_WIN_SIZE); cnt0+cnt1 > 0 {
		errPct = float64(cnt1*100) / float64(cnt0+cnt1)
	}
	return int(errPct + 0.5)
}

func (cb *circuitBreaker) allowRequest() (allow, singletest bool) {
	if cb.isForceOpen {
		return false, false
	}

	nsec := timex.UnixNano()
	if cb.isOpen.Load() {
		// recovery?
		if cb.recovery.IsHealthy() {
			cb.trigger(false)
			return true, false
		}

		// single test should be applied when circuit is opened
		lastTestAt := cb.lastTestAt.Load()
		if nsec >= lastTestAt+cb.KeepAliveInterval.Nanoseconds() {
			swapped := cb.lastTestAt.CompareAndSwap(lastTestAt, nsec)
			if swapped {
				log.Printf("hystrix-go: allowing single test.\n")
			}
			return swapped, true
		}

		return false, false
	}

	// regardless of whether it succeeds or not, do not perform a health check when the current qps is too low.
	if int(cb.qps(nsec)) < cb.MinQPSThreshold {
		return true, false
	}

	// too many failures, open the circuitbreaker
	if cb.failRate(nsec) >= cb.FailureThreshold {
		cb.trigger(true)
		return false, false
	}

	return true, false
}

func (cb *circuitBreaker) trigger(open bool) {
	if open {
		if cb.isOpen.CompareAndSwap(false, true) {
			cb.lastTestAt.Store(timex.UnixNano())
			cb.statistic.Reset()
			log.Printf("hystrix-go: opening circuitbreaker.\n")
		}
	} else {
		if cb.isOpen.CompareAndSwap(true, false) {
			cb.recovery.Reset()
			cb.statistic.Reset()
			log.Printf("hystrix-go: closing circuitbreaker.\n")
		}
	}
}

type snapshot struct {
	err       error
	isTestReq bool
	startNano int64 // nanoseconds
	stopNano  int64
}

func (cb *circuitBreaker) feedback(s snapshot) {
	if cb.isOpen.Load() && s.isTestReq {
		cb.recovery.Update(s.err == nil)
	}

	nsec := s.stopNano
	if s.err == nil {
		cb.statistic.At(nsec).DualIncrBy(1, 0)
	} else {
		cb.statistic.At(nsec).DualIncrBy(0, 1)
	}
}

type singleTestMeta struct{}

// GoC runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func (cb *circuitBreaker) GoC(ctx context.Context, runable func(context.Context) error) chan error {
	nanosec := timex.UnixNano()
	errchan := make(chan error, 1)
	if !cb.tickets.TryAcquire(1) {
		cb.feedback(snapshot{ErrMaxConcurrency, false, nanosec, nanosec})
		errchan <- ErrMaxConcurrency
		return errchan
	}

	allow, singletest := cb.allowRequest()
	if !allow {
		cb.feedback(snapshot{ErrCircuitOpen, false, nanosec, nanosec})
		errchan <- ErrCircuitOpen
		return errchan
	}

	resultChan := make(chan snapshot, 1)
	runStart := timex.UnixNano()
	go func() {
		err := runable(context.WithValue(ctx, singleTestMeta{}, singletest))
		resultChan <- snapshot{err, singletest, runStart, timex.UnixNano()}
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
			cb.feedback(snapshot{ErrCanceled, singletest, runStart, timex.UnixNano()})

		case <-tmoCtx.Done():
			errchan <- ErrTimeout
			cb.feedback(snapshot{ErrTimeout, singletest, runStart, timex.UnixNano()})
		}
	}()

	return errchan
}
