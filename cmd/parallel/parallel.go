package parallel

import (
	"context"
	"errors"
	"sync"

	"github.com/avast/retry-go/v4"
)

const (
	defaultAttempts    = 1
	defaultConcurrency = 4
)

type Option func(*Group)

// 0 for infinate retry until success
func Attempts(v int) Option {
	return func(g *Group) {
		g.attempts = v
	}
}

func StopOnError() Option {
	return func(g *Group) {
		g.stopOnError = true
	}
}

func StopOnSuccess() Option {
	return func(g *Group) {
		g.stopOnSuccess = true
	}
}

func CancelPending() Option {
	return func(g *Group) {
		g.cancelPending = true
	}
}

func MaxConcurrency(v int) Option {
	return func(g *Group) {
		g.maxConcurrency = v
	}
}

type Group struct {
	ctx            context.Context
	cancel         context.CancelFunc
	attempts       int // try attempts for every func
	maxConcurrency int // run concurrency
	stopOnError    bool
	stopOnSuccess  bool
	cancelPending  bool
}

func WithContext(opt ...Option) Group {
	ctx, cancel := context.WithCancel(context.Background())
	g := Group{
		ctx:            ctx,
		cancel:         cancel,
		attempts:       defaultAttempts,
		maxConcurrency: defaultConcurrency,
		stopOnError:    false,
		stopOnSuccess:  false,
		cancelPending:  false,
	}
	for _, f := range opt {
		f(&g)
	}
	return g
}

func (g *Group) RunParallel(fns ...func(context.Context) error) error {
	wg := sync.WaitGroup{}
	wg.Add(g.maxConcurrency)

	finishCtx, onFinish := context.WithCancel(context.Background())
	defer onFinish()

	works := make(chan func(context.Context) error, len(fns))
	defer close(works)

	for _, f := range fns {
		works <- f
	}

	forceSuccess := false
	errslice := []error{}
	for i := 0; i < g.maxConcurrency; i++ {
		go func() {
			defer wg.Done()
			select {
			case <-g.ctx.Done():
				return
			case <-finishCtx.Done():
				return
			case f := <-works:
				if err := retry.Do(func() error {
					return f(g.ctx)
				}, retry.Attempts(uint(g.attempts))); err == nil {
					if g.stopOnSuccess {
						onFinish()
						if g.cancelPending {
							g.cancel()
						}
						forceSuccess = true
						return
					}
				} else {
					errslice = append(errslice, err)
					if g.stopOnError {
						onFinish()
						if g.cancelPending {
							g.cancel()
						}
						return
					}
				}
			}
		}()
	}
	wg.Wait()

	if forceSuccess || len(errslice) == 0 {
		return nil
	}
	return errors.Join(errslice...)
}