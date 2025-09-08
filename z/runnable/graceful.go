package runnable

import (
	"context"
	"errors"
	"time"
)

type Graceful struct {
	Start func(context.Context) error
	Stop  func() error
}

func (g *Graceful) GoContext(ctx context.Context) error {
	resultCh := make(chan error, 1)
	if g.Stop == nil {
		g.Stop = func() error { return nil }
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		go func() { resultCh <- g.Start(ctx) }()
	}
	return errors.Join(<-resultCh, g.Stop())
}

func (g *Graceful) Go() error {
	return g.GoContext(context.Background())
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
