package cmd

import (
	"context"
	"time"
)

func Delayed(ctx context.Context, delay time.Duration, f func(context.Context)) (context.Context, context.CancelFunc) {
	cancelCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		select {
		case <-time.After(delay):
			f(cancelCtx)
		case <-cancelCtx.Done():
		}
	}()
	return cancelCtx, cancel
}

func Go(f func(ctx context.Context)) (context.Context, context.CancelFunc) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	async(func() {
		f(cancelCtx)
		cancel()
	})
	return cancelCtx, cancel
}

func Async(f func()) {
	async(f)
}

func async(f func()) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	go func() {
		cancel()
		f()
	}()
	<-cancelCtx.Done()
}

func Timed(ctx context.Context, tmo time.Duration, f func(context.Context)) {
	ctx, cancel := context.WithTimeout(ctx, tmo)
	defer cancel()
	f(ctx)
}
