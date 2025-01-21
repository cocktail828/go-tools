package cmd

import (
	"context"
	"errors"
	"time"
)

type Graceful struct {
	Start func() error
	Stop  func() error
}

func (g *Graceful) Do(ctx context.Context) error {
	runningCtx, cancel := context.WithCancelCause(ctx)
	go func() {
		cancel(g.Start())
	}()

	<-runningCtx.Done()
	return errors.Join(g.Stop(), context.Cause(runningCtx))
}

func Timed(d time.Duration, f func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.TODO(), d)
	defer cancel()

	c := make(chan error, 1)
	go func() { c <- f(ctx) }()

	select {
	case <-ctx.Done():
		return context.DeadlineExceeded
	case err := <-c:
		return err
	}
}
