package hystrix

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

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

type Hystrix struct {
	Config
	lastTestAt    atomic.Int64 // last single test timestamp in nanoseconds
	isOpen        atomic.Bool  // circuit state
	isForceOpen   bool         // manually turn on/off the circuit
	assigner      *Assigner    // for concurrency control
	statistic     *rolling.DualRolling
	recovery      *recovery
	logger        xlog.Printer
	stateChangeAt atomic.Int64 // state change timestamp in nanoseconds
}

func NewHystrix(cfg Config) *Hystrix {
	h := &Hystrix{
		Config:    cfg,
		assigner:  &Assigner{maxCount: cfg.MaxConcurrency.Val.Get()},
		statistic: rolling.NewRolling(128).Dual(),
		recovery:  newRecovery(cfg.KeepAliveProbes.Val.Get()),
		logger:    xlog.NopPrinter{},
	}
	h.MaxConcurrency.OnUpdate.Set(func(v int) { h.assigner.Resize(v) })
	h.stateChangeAt.Store(timex.UnixNano())

	return h
}

type Statistic struct {
	Concurrency   int     // 并发数
	FailRate      float64 // 0~100
	QPS           float64 // 3s 内的 QPS
	StateDuration float64 // 熔断器当前状态持续时间（毫秒）
	IsOpen        bool    // 熔断器是否开启
}

func (h *Hystrix) Statistic() Statistic {
	nsec := timex.UnixNano()
	stateDuration := float64(nsec-h.stateChangeAt.Load()) / float64(time.Millisecond)

	return Statistic{
		Concurrency:   h.assigner.Allocated(),
		FailRate:      h.failRate(nsec),
		QPS:           h.qps(nsec),
		StateDuration: stateDuration,
		IsOpen:        h.isOpen.Load(),
	}
}

func (s Statistic) String() string {
	state := "closed"
	if s.IsOpen {
		state = "open"
	}
	return fmt.Sprintf("{State: %s, Concurrency: %d, FailRate: %.2f%%, QPS: %.2f, Duration: %.1fms}",
		state, s.Concurrency, s.FailRate, s.QPS, s.StateDuration)
}

// open or close the circuitbreaker manually
func (h *Hystrix) Trigger(isopen bool) {
	h.isForceOpen = isopen
}

func (h *Hystrix) qps(nsec int64) float64 {
	v0, v1 := h.statistic.At(nsec).QPS(_COUNTER_WIN_SIZE)
	return v0 + v1
}

func (h *Hystrix) failRate(nsec int64) float64 {
	var errPct float64
	if cnt0, cnt1, _ := h.statistic.At(nsec).Count(_COUNTER_WIN_SIZE); cnt0+cnt1 > 0 {
		errPct = float64(cnt1*100) / float64(cnt0+cnt1)
	}
	return errPct
}

func (h *Hystrix) allowRequest() (allow, singletest bool) {
	if h.isForceOpen {
		return false, false
	}

	nsec := timex.UnixNano()
	if h.isOpen.Load() {
		// recovery?
		// Only check recovery status if we have enough probe data
		if h.recovery.IsHealthy() {
			h.trigger(false)
			return true, false
		}

		// single test should be applied when circuit is opened
		lastTestAt := h.lastTestAt.Load()
		keepAliveInterval := h.KeepAliveInterval.Val.Get().Nanoseconds()

		if nsec >= lastTestAt+keepAliveInterval {
			swapped := h.lastTestAt.CompareAndSwap(lastTestAt, nsec)
			if swapped {
				h.logger.Printf("hystrix-go: allowing single test.")
				return true, true
			}
			return false, false
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

func (h *Hystrix) trigger(open bool) {
	if open {
		if h.isOpen.CompareAndSwap(false, true) {
			now := timex.UnixNano()
			h.stateChangeAt.Store(now)
			h.lastTestAt.Store(now)
			h.logger.Printf("hystrix-go: opening circuitbreaker. Stats: %s", h.Statistic())
			h.statistic.Reset()
		}
	} else {
		if h.isOpen.CompareAndSwap(true, false) {
			h.stateChangeAt.Store(timex.UnixNano())
			h.logger.Printf("hystrix-go: closing circuitbreaker. Recovery status: %s", h.recovery)
			h.recovery.Reset()
		}
	}
}

type snapshot struct {
	err       error
	isTestReq bool
	startNano int64 // nanoseconds
	stopNano  int64
}

func (h *Hystrix) feedback(s snapshot) {
	nsec := s.stopNano
	latency := float64(nsec-s.startNano) / float64(time.Millisecond)

	if h.isOpen.Load() && s.isTestReq {
		h.recovery.Update(s.err == nil)
		// record single test request result
		if s.err == nil {
			h.logger.Printf("hystrix-go: single test request succeeded. Latency: %.2fms. Recovery status: %s",
				latency, h.recovery)
		} else {
			h.logger.Printf("hystrix-go: single test request failed: %v. Latency: %.2fms. Recovery status: %s",
				s.err, latency, h.recovery)
		}
	}

	// record circuitbreaker state change
	currentErrRate := h.failRate(nsec)
	threshold := float64(h.FailureThreshold.Val.Get())
	shouldOpen := currentErrRate >= threshold

	if h.isOpen.Load() != shouldOpen {
		h.logger.Printf("hystrix-go: circuit state change detected - Current: %v, ShouldOpen: %v, ErrorRate: %.2f%%, Threshold: %.0f%%, QPS: %.2f",
			h.isOpen.Load(), shouldOpen, currentErrRate, threshold, h.qps(nsec))
	}

	// update statistics
	if s.err == nil {
		h.statistic.At(nsec).IncrBy(1, 0)
	} else {
		h.statistic.At(nsec).IncrBy(0, 1)
	}
}

type singleTestMeta struct{}

// GoC runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func (h *Hystrix) GoC(ctx context.Context, runable func(context.Context) error) chan error {
	nanosec := timex.UnixNano()
	errchan := make(chan error, 1)

	allow, singletest := h.allowRequest()
	if !allow {
		h.feedback(snapshot{ErrCircuitOpen, false, nanosec, nanosec})
		errchan <- ErrCircuitOpen
		return errchan
	}

	if !h.assigner.TryAcquire() {
		h.feedback(snapshot{ErrMaxConcurrency, false, nanosec, nanosec})
		errchan <- ErrMaxConcurrency
		return errchan
	}

	resultChan := make(chan snapshot, 1)
	runStart := timex.UnixNano()
	go func() {
		err := runable(context.WithValue(ctx, singleTestMeta{}, singletest))
		resultChan <- snapshot{err, singletest, runStart, timex.UnixNano()}
	}()

	go func() {
		defer func() {
			h.assigner.Release()
			if r := recover(); r != nil {
				h.logger.Printf("hystrix-go: panic when waitting execute result: %v", r)
			}
		}()

		tmoCtx, cancel := context.WithTimeout(context.Background(), h.Timeout.Val.Get())
		defer cancel()

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
func (h *Hystrix) Go(name string, runnable func() error) chan error {
	return h.GoC(
		context.TODO(),
		func(_ context.Context) error { return runnable() },
	)
}

// Do runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func (h *Hystrix) Do(name string, runnable func() error) error {
	return <-h.GoC(
		context.TODO(),
		func(_ context.Context) error { return runnable() },
	)
}

// DoC runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func (h *Hystrix) DoC(ctx context.Context, name string, runnable func(context.Context) error) error {
	return <-h.GoC(
		ctx,
		func(ctx context.Context) error { return runnable(ctx) },
	)
}
