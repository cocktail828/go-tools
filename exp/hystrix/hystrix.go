package hystrix

import (
	"context"
	"errors"
	"sync"
	_ "unsafe"
)

var (
	circuitBreakerMap sync.Map
)

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

func GetCircuit(name string) *circuitBreaker {
	var cb *circuitBreaker
	if val, ok := circuitBreakerMap.Load(name); ok {
		cb = val.(*circuitBreaker)
	} else {
		cb = NewCircuitBreaker(Config{})
		circuitBreakerMap.Store(name, cb)
	}

	return cb
}

func Configure(name string, cfg Config) {
	GetCircuit(name).Update(cfg)
}

// GoC runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func GoC(ctx context.Context, name string, runnable func(context.Context) error) chan error {
	return GetCircuit(name).GoC(ctx, runnable)
}

// Go runs your function while tracking the health of previous calls to it.
// If your function begins slowing down or failing repeatedly, we will block
// new calls to it for you to give the dependent service time to repair.
//
// Define a fallback function if you want to define some code to execute during outages.
func Go(name string, runnable func() error) chan error {
	return GoC(
		context.TODO(),
		name,
		func(_ context.Context) error { return runnable() },
	)
}

// Do runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func Do(name string, runnable func() error) error {
	return <-GoC(
		context.TODO(),
		name,
		func(_ context.Context) error { return runnable() },
	)
}

// DoC runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func DoC(ctx context.Context, name string, runnable func(context.Context) error) error {
	return <-GoC(
		ctx,
		name,
		func(ctx context.Context) error { return runnable(ctx) },
	)
}
