package cmd

import (
	"context"
	"time"
)

func delayedContext(ctx context.Context, delay time.Duration, f func(context.Context)) (context.Context, context.CancelFunc) {
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

func DelayedContext(ctx context.Context, delay time.Duration, f func(context.Context)) (context.Context, context.CancelFunc) {
	return delayedContext(ctx, delay, f)
}

func Delayed(delay time.Duration, f func(context.Context)) (context.Context, context.CancelFunc) {
	return delayedContext(context.Background(), delay, f)
}

func goContext(ctx context.Context, f func(ctx context.Context)) (context.Context, context.CancelFunc) {
	cancelCtx, cancel := context.WithCancel(ctx)
	async(func() {
		f(cancelCtx)
		cancel()
	})
	return cancelCtx, cancel
}

func GoContext(ctx context.Context, f func(ctx context.Context)) (context.Context, context.CancelFunc) {
	return goContext(ctx, f)
}

func Go(f func(ctx context.Context)) (context.Context, context.CancelFunc) {
	return goContext(context.Background(), f)
}

func async(f func()) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	go func() {
		cancel()
		f()
	}()
	<-cancelCtx.Done()
}

func Async(f func()) {
	async(f)
}

func timedContext(ctx context.Context, tmo time.Duration, f func(context.Context)) {
	ctx, cancel := context.WithTimeout(ctx, tmo)
	defer cancel()
	f(ctx)
}

func TimedContext(ctx context.Context, tmo time.Duration, f func(context.Context)) {
	timedContext(ctx, tmo, f)
}

func Timed(tmo time.Duration, f func(context.Context)) {
	timedContext(context.Background(), tmo, f)
}
