package hystrix

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

type Now func() time.Time

var (
	nowFunc = time.Now // for debug
)

func SetNow(f Now) { nowFunc = f }

type runnable func() error
type runnableCtx func(context.Context) error

var (
	// ErrMaxConcurrency occurs when too many of the same named command are executed at the same time.
	ErrMaxConcurrency = errors.New("hystrix: max concurrency")
	// ErrCircuitOpen returns when an execution attempt "short circuits". This happens due to the circuit being measured as unhealthy.
	ErrCircuitOpen = errors.New("hystrix: circuit open")
	// ErrTimeout occurs when the provided function takes too long to execute.
	ErrTimeout = errors.New("hystrix: the operation is timeout")
	// ErrTimeout occurs when the provided function takes too long to execute.
	ErrCanceled = errors.New("hystrix: the operation is canceled")
)

// Do runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func Do(name string, run runnable) error {
	return <-GoC(context.Background(), name, func(_ context.Context) error { return run() })
}

// DoC runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func DoC(ctx context.Context, name string, run runnableCtx) error {
	return <-GoC(ctx, name, func(ctx context.Context) error { return run(ctx) })
}

// Go runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func Go(name string, run runnable) chan error {
	return GoC(context.Background(), name, func(_ context.Context) error { return run() })
}

type result struct {
	err     error
	elapsed time.Duration // command execute cost time
}

// GoC runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func GoC(ctx context.Context, name string, run runnableCtx) chan error {
	startAt := nowFunc()
	circuit := GetCircuit(name)
	var holdTicket atomic.Bool

	resultChan := make(chan result, 1)
	go func() {
		// Circuits get opened when recent executions have shown to have a high error rate.
		// Rejecting new executions allows backends to recover, and the circuit will allow
		// new traffic when it feels a healthly state has returned.
		if !circuit.allowRequest() {
			resultChan <- result{err: ErrCircuitOpen}
			return
		}

		// As backends falter, requests take longer but don't always fail.
		//
		// When requests slow down but the incoming rate of requests stays the same, you have to
		// run more at a time to keep up. By controlling concurrency during these situations, you can
		// shed load which accumulates due to the increasing ratio of active commands to incoming requests.
		select {
		case <-circuit.tickets:
		default:
			resultChan <- result{err: ErrMaxConcurrency}
			return
		}

		holdTicket.Store(true)
		defer func() {
			if holdTicket.CompareAndSwap(true, false) {
				circuit.tickets <- struct{}{}
			}
		}()

		runStart := nowFunc()
		err := run(ctx)
		resultChan <- result{
			err:     err,
			elapsed: time.Since(runStart),
		}
	}()

	errChan := make(chan error, 1)
	go func() {
		tmoCtx, cancel := context.WithTimeout(context.Background(), circuit.setting.Timeout)
		defer cancel()
		defer func() {
			if holdTicket.CompareAndSwap(true, false) {
				circuit.tickets <- struct{}{}
			}
		}()

		select {
		case result := <-resultChan:
			var ev EventType
			switch result.err {
			case nil:
				ev = SuccessEvent
			case ErrCircuitOpen:
				ev = ShortCircuitEvent
			case ErrMaxConcurrency:
				ev = MaxConcurrencyEvent
			default:
				ev = ErrorEvent
			}
			errChan <- result.err
			circuit.feedback(ev, startAt, nowFunc(), result.elapsed)

		case <-ctx.Done():
			errChan <- ErrCanceled
			circuit.feedback(CancelEvent, startAt, nowFunc(), time.Since(startAt))
		case <-tmoCtx.Done():
			errChan <- ErrTimeout
			circuit.feedback(TimeoutEvent, startAt, nowFunc(), time.Since(startAt))
		}
	}()

	return errChan
}
