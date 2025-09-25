package hystrix

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/cocktail828/go-tools/algo/rolling"
	"github.com/cocktail828/go-tools/xlog"
	"github.com/cocktail828/go-tools/z/timex"
)

var (
	// ErrMaxConcurrency occurs when too many of the same named command are executed at the same time.
	ErrMaxConcurrency = errors.New("hystrix: max concurrency reached")
	// ErrCircuitOpen returns when an execution attempt "short circuits". This happens due to the circuit being measured as unhealthy.
	ErrCircuitOpen = errors.New("hystrix: circuit is open")
	// ErrTimeout occurs when the provided function takes too long to execute.
	ErrTimeout = errors.New("hystrix: the operation is timeout")
	// ErrTimeout occurs when the provided function takes too long to execute.
	ErrCanceled = errors.New("hystrix: the operation is canceled")
)

const (
	_COUNTER_WIN_SIZE = 24 // 128ms*24=3.072s
)

type hystrix struct {
	Config
	lastTestAt  atomic.Int64 // last single test timestamp in nanoseconds
	isOpen      atomic.Bool  // circuit state
	isForceOpen bool         // manually turn on/off the circuit
	assigner    *Assigner    // for concurrency control
	statistic   *rolling.Rolling
	recovery    recovery
	logger      xlog.Logger
}

func NewHystrix(cfg Config, logger xlog.Logger) *hystrix {
	if logger == nil {
		logger = xlog.NoopLogger{}
	}

	h := &hystrix{
		Config:    cfg,
		assigner:  &Assigner{maxCount: cfg.MaxConcurrency.Val.Get()},
		statistic: rolling.NewRolling(128),
		recovery:  recovery{array: make([]bool, cfg.KeepAliveProbes.Val.Get())},
		logger:    logger,
	}
	h.MaxConcurrency.OnUpdate.Set(func(v int) { h.assigner.Resize(v) })

	return h
}

type Statistic struct {
	Concurrency int     // 并发数
	FailRate    float64 // 0~100
	QPS         float64 // 3s 内的 QPS
}

func (h *hystrix) Statistic() Statistic {
	nsec := timex.UnixNano()
	return Statistic{
		Concurrency: h.assigner.Allocated(),
		FailRate:    h.failRate(nsec),
		QPS:         h.qps(nsec),
	}
}

// open or close the circuitbreaker manually
func (h *hystrix) Trigger(isopen bool) {
	h.isForceOpen = isopen
}

func (h *hystrix) qps(nsec int64) float64 {
	v0, v1 := h.statistic.At(nsec).DualQPS(_COUNTER_WIN_SIZE)
	return v0 + v1
}

func (h *hystrix) failRate(nsec int64) float64 {
	var errPct float64
	if cnt0, cnt1, _ := h.statistic.At(nsec).DualCount(_COUNTER_WIN_SIZE); cnt0+cnt1 > 0 {
		errPct = float64(cnt1*100) / float64(cnt0+cnt1)
	}
	return errPct + 0.5
}

func (h *hystrix) allowRequest() (allow, singletest bool) {
	if h.isForceOpen {
		return false, false
	}

	nsec := timex.UnixNano()
	if h.isOpen.Load() {
		// recovery?
		if h.recovery.IsHealthy() {
			h.trigger(false)
			return true, false
		}

		// single test should be applied when circuit is opened
		lastTestAt := h.lastTestAt.Load()
		if nsec >= lastTestAt+h.KeepAliveInterval.Val.Get().Nanoseconds() {
			swapped := h.lastTestAt.CompareAndSwap(lastTestAt, nsec)
			if swapped {
				h.logger.Debugln("hystrix-go: allowing single test.")
			}
			return swapped, true
		}

		return false, false
	}

	// regardless of whether it succeeds or not, do not perform a health check when the current qps is too low.
	if int(h.qps(nsec)) < h.MinQPSThreshold.Val.Get() {
		return true, false
	}

	// too many failures, open the circuitbreaker
	if int(h.failRate(nsec)) >= h.FailureThreshold.Val.Get() {
		h.trigger(true)
		return false, false
	}

	return true, false
}

func (h *hystrix) trigger(open bool) {
	if open {
		if h.isOpen.CompareAndSwap(false, true) {
			h.lastTestAt.Store(timex.UnixNano())
			h.statistic.Reset()
			h.logger.Debugln("hystrix-go: opening circuitbreaker.")
		}
	} else {
		if h.isOpen.CompareAndSwap(true, false) {
			h.recovery.Reset()
			h.statistic.Reset()
			h.logger.Debugln("hystrix-go: closing circuitbreaker.")
		}
	}
}

type snapshot struct {
	err       error
	isTestReq bool
	startNano int64 // nanoseconds
	stopNano  int64
}

func (h *hystrix) feedback(s snapshot) {
	if h.isOpen.Load() && s.isTestReq {
		h.recovery.Update(s.err == nil)
	}

	nsec := s.stopNano
	if s.err == nil {
		h.statistic.At(nsec).DualIncrBy(1, 0)
	} else {
		h.statistic.At(nsec).DualIncrBy(0, 1)
	}
}

type singleTestMeta struct{}

// GoC runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func (h *hystrix) GoC(ctx context.Context, runable func(context.Context) error) chan error {
	nanosec := timex.UnixNano()
	errchan := make(chan error, 1)
	if !h.assigner.TryAcquire() {
		h.feedback(snapshot{ErrMaxConcurrency, false, nanosec, nanosec})
		errchan <- ErrMaxConcurrency
		return errchan
	}

	allow, singletest := h.allowRequest()
	if !allow {
		h.feedback(snapshot{ErrCircuitOpen, false, nanosec, nanosec})
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
		tmoCtx, cancel := context.WithTimeout(context.Background(), h.Timeout.Val.Get())
		defer cancel()
		defer h.assigner.Release()

		select {
		case rc := <-resultChan:
			errchan <- rc.err
			h.feedback(rc)

		case <-ctx.Done():
			errchan <- ErrCanceled
			h.feedback(snapshot{ErrCanceled, singletest, runStart, timex.UnixNano()})

		case <-tmoCtx.Done():
			errchan <- ErrTimeout
			h.feedback(snapshot{ErrTimeout, singletest, runStart, timex.UnixNano()})
		}
	}()

	return errchan
}

// Go runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func (h *hystrix) Go(name string, runnable func() error) chan error {
	return h.GoC(
		context.TODO(),
		func(_ context.Context) error { return runnable() },
	)
}

// Do runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func (h *hystrix) Do(name string, runnable func() error) error {
	return <-h.GoC(
		context.TODO(),
		func(_ context.Context) error { return runnable() },
	)
}

// DoC runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func (h *hystrix) DoC(ctx context.Context, name string, runnable func(context.Context) error) error {
	return <-h.GoC(
		ctx,
		func(ctx context.Context) error { return runnable(ctx) },
	)
}
