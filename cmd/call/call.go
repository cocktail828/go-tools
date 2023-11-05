package call

import (
	"context"
	"time"
)

func DelayedContext(ctx context.Context, delay time.Duration, f func(context.Context) error) (context.CancelFunc, <-chan error) {
	cancelCtx, cancel := context.WithCancel(ctx)
	errChan := make(chan error, 2)
	go func() {
		select {
		case <-time.After(delay):
			errChan <- f(cancelCtx)
		case <-cancelCtx.Done():
			errChan <- context.Canceled
		}
	}()
	return cancel, errChan
}

func Delayed(delay time.Duration, f func(context.Context) error) (context.CancelFunc, <-chan error) {
	return DelayedContext(context.Background(), delay, f)
}

func TimedContext(ctx context.Context, tmo time.Duration, f func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, tmo)
	defer cancel()
	return f(ctx)
}

func Timed(tmo time.Duration, f func(context.Context) error) error {
	return TimedContext(context.Background(), tmo, f)
}
