package parallel

import (
	"context"
	"fmt"
	"sync"
)

type Event int

const (
	None    Event = iota
	Failure Event = iota
	Success Event = iota
)

type Option func(*Group)

func AbortWhen(e Event) Option {
	return func(g *Group) {
		g.abortEvent = e
	}
}

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid, has no limit on the number of active goroutines,
// and does not cancel on error.
type Group struct {
	ctx        context.Context
	cancel     context.CancelCauseFunc
	wg         sync.WaitGroup
	token      chan struct{}
	once       sync.Once
	err        error
	abortEvent Event
}

func (g *Group) done() {
	if g.token != nil {
		<-g.token
	}
	g.wg.Done()
}

// WithContext returns a new Group and an associated Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func WithContext(ctx context.Context, opts ...Option) (*Group, context.Context) {
	subctx, cancel := context.WithCancelCause(ctx)
	g := &Group{ctx: subctx, cancel: cancel}
	for _, f := range opts {
		f(g)
	}
	return g, ctx
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}

func (g *Group) call(f func() error) {
	select {
	case <-g.ctx.Done():
		return
	default:
		err := f()
		if (err != nil && g.abortEvent == Failure) ||
			(err == nil && g.abortEvent == Success) {
			g.once.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(g.err)
				}
			})
		}
	}
}

// Go calls the given function in a new goroutine.
// It blocks until the new goroutine can be added without the number of
// active goroutines in the group exceeding the configured limit.
//
// The first call to return a non-nil error cancels the group's context, if the
// group was created by calling WithContext. The error will be returned by Wait.
func (g *Group) Go(f func() error) {
	if g.token != nil {
		g.token <- struct{}{}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()
		g.call(f)
	}()
}

// TryGo calls the given function in a new goroutine only if the number of
// active goroutines in the group is currently below the configured limit.
//
// The return value reports whether the goroutine was started.
func (g *Group) TryGo(f func() error) bool {
	if g.token != nil {
		select {
		case g.token <- struct{}{}:
			// Note: this allows barging iff channels in general allow barging.
		default:
			return false
		}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()
		g.call(f)
	}()
	return true
}

// SetLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
//
// The limit must not be modified while any goroutines in the group are active.
func (g *Group) SetLimit(n int) {
	if n < 0 {
		g.token = nil
		return
	}
	if len(g.token) != 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the group are still active", len(g.token)))
	}
	g.token = make(chan struct{}, n)
}
