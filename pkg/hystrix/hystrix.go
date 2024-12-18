package hystrix

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type runFn func() error
type fallbackFn func(error) error
type runFnCtx func(context.Context) error
type fallbackFnCtx func(context.Context, error) error

var (
	ErrMaxConcurrency = errors.New("circuit: max concurrency")
	ErrCircuitOpen    = errors.New("circuit: circuit is opened")
)

// Go runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func Go(name string, run runFn, fallback fallbackFn) chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- DoC(context.Background(), name,
			func(_ context.Context) error { return run() },
			func(_ context.Context, err error) error {
				if fallback != nil {
					return fallback(err)
				}
				return nil
			})
	}()
	return errCh
}

// GoC runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func GoC(ctx context.Context, name string, run runFnCtx, fallback fallbackFnCtx) chan error {
	errCh := make(chan error, 1)
	go func() {
		if err := DoC(ctx, name, run, fallback); err != nil {
			errCh <- err
		}
	}()
	return errCh
}

// Do runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func Do(name string, run runFn, fallback fallbackFn) error {
	return DoC(context.Background(), name,
		func(_ context.Context) error { return run() },
		func(_ context.Context, err error) error {
			if fallback != nil {
				return fallback(err)
			}
			return nil
		})
}

// DoC runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func DoC(ctx context.Context, name string, run runFnCtx, fallback fallbackFnCtx) error {
	cmd := command{
		start: time.Now(),
		runFn: run,
		fbFn:  fallback,
	}
	cb := getCircuit(name)
	defer cmd.report(cb)

	// Circuits get opened when recent executions have shown to have a high error rate.
	// Rejecting new executions allows backends to recover, and the circuit will allow
	// new traffic when it feels a healthly state has returned.
	if !cb.AllowRequest() {
		cmd.fallback(ctx, ErrCircuitOpen)
		return cmd.getError()
	}

	if !cb.TryAcquire() {
		cmd.fallback(ctx, ErrMaxConcurrency)
		return cmd.getError()
	}
	defer cb.Release()

	finishChan := make(chan struct{})
	ctx, cancel := context.WithTimeout(ctx, cb.Timeout)
	defer cancel()
	go func() {
		defer close(finishChan)
		err := run(ctx)
		cmd.elapsed = time.Since(cmd.start)
		if err != nil {
			cmd.fallback(ctx, err)
		}
	}()

	select {
	case <-finishChan:
	case <-ctx.Done():
		cmd.fallback(ctx, ctx.Err())
	}

	return cmd.getError()
}
