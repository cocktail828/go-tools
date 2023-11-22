package cmd

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/avast/retry-go/v4"
	"github.com/cocktail828/go-tools/cmd/event"
	"github.com/cocktail828/go-tools/z/reflectx"
	"github.com/pkg/errors"
)

var (
	ErrUserCancel    = errors.Errorf("user cancel the parallel group")
	ErrAlreadyInited = errors.Errorf("parallel group already started")
)

const (
	defaultAttempts    = 1
	defaultConcurrency = 4
)

type ParallelOption func(*ParallelGroup)

// 0 for infinate retry until success
func Attempts(v int) ParallelOption {
	return func(g *ParallelGroup) {
		g.attempts = v
	}
}

func AbortEvent(event event.Event) ParallelOption {
	return func(g *ParallelGroup) {
		g.abortEvent = event
	}
}

func MaxConcurrency(v int) ParallelOption {
	return func(g *ParallelGroup) {
		g.concurrency = v
	}
}

func MultiOption(opts ...ParallelOption) ParallelOption {
	return func(g *ParallelGroup) {
		for _, f := range opts {
			f(g)
		}
	}
}

type ParallelGroup struct {
	inited      atomic.Bool
	ctx         context.Context
	cancel      context.CancelCauseFunc
	attempts    int // try attempts for every func
	concurrency int // run concurrency
	abortEvent  event.Event
}

func (g *ParallelGroup) Stop() {
	if g.inited.Load() {
		g.cancel(ErrUserCancel)
	}
}

func (g *ParallelGroup) Run(opt ParallelOption, fns ...func(context.Context) error) error {
	if len(fns) == 0 {
		return nil
	}
	wc := make(chan func(context.Context) error, len(fns))
	for _, f := range fns {
		wc <- f
	}
	return g.RunWithChan(opt, wc)
}

func (g *ParallelGroup) RunWithChan(opt ParallelOption, fnChan chan func(context.Context) error) error {
	if !g.inited.CompareAndSwap(false, true) {
		return ErrAlreadyInited
	}
	g.ctx, g.cancel = context.WithCancelCause(context.Background())
	g.attempts = defaultAttempts
	g.concurrency = defaultConcurrency
	g.abortEvent = event.None
	opt(g)

	wg := sync.WaitGroup{}
	wg.Add(g.concurrency)
	var ferr error
	for i := 0; i < g.concurrency; i++ {
		go func() {
			defer wg.Done()
			select {
			case <-g.ctx.Done():
				return
			case f := <-fnChan:
				err := retry.Do(func() error {
					return f(g.ctx)
				}, retry.Attempts(uint(g.attempts)))

				if reflectx.IsNil(err) {
					if g.abortEvent == event.Success {
						g.cancel(nil)
						return
					}
				} else {
					ferr = err
					if g.abortEvent == event.Error {
						g.cancel(err)
						return
					}
				}
			}
		}()
	}
	wg.Wait()

	switch g.abortEvent {
	case event.Success:
		if err := g.ctx.Err(); reflectx.IsNil(err) {
			return nil
		}
	case event.Error:
		if err := g.ctx.Err(); !reflectx.IsNil(err) {
			return err
		}
	}
	return ferr
}
