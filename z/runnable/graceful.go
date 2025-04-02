package runnable

import (
	"context"
	"errors"
	"time"
)

type Graceful struct {
	Start func() error
	Stop  func() error
}

func (g *Graceful) Launch(pctx context.Context) error {
	ctx, cancel := context.WithCancelCause(pctx)
	go func() { cancel(g.Start()) }()

	if g.Stop == nil {
		g.Stop = func() error { return nil }
	}

	<-ctx.Done()
	err := context.Cause(ctx)
	if err == context.Canceled {
		err = nil
	}
	return errors.Join(err, g.Stop())
}

func Timeout(d time.Duration, f func(context.Context) error) error {
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
