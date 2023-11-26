package retry

import (
	"context"
	"time"

	"github.com/cocktail828/go-tools/z"
	"github.com/cocktail828/go-tools/z/retry/backoff"
	"github.com/pkg/errors"
)

func Do(f func(context.Context) error, opts ...Option) error {
	_, err := DoAny(func(ctx context.Context) (any, error) {
		return nil, f(ctx)
	}, opts...)
	return err
}

func DoAny[T any](f func(context.Context) (T, error), opts ...Option) (T, error) {
	cfg := &config{
		attempts:        0,
		attemptForError: make(map[error]struct{}),
		backoff:         &backoff.Fixed{Value: 100},
		lastErrorOnly:   false,
		ctx:             context.Background(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	var emptyT T
	errs := z.Error{}
	fn := func() (T, bool, error) {
		defer func() {
			if err := recover(); err != nil {
				errs = append(errs, errors.Errorf("recover: %v", err))
			}
		}()
		select {
		case <-cfg.ctx.Done():
			errs = append(errs, cfg.ctx.Err())
			return emptyT, false, cfg.ctx.Err()
		default:
			t, e := f(cfg.ctx)
			if e != nil {
				errs = append(errs, e)
				return emptyT, true, e
			}
			return t, false, e
		}
	}

	n := 0
loop:
	for {
		t, c, e := fn()
		if e == nil {
			return t, nil
		}
		if !c {
			break
		}

		n++
		shouldAttempt := (cfg.attempts <= 0 || n < cfg.attempts) &&
			func() bool {
				if cfg.retryIf != nil {
					return cfg.retryIf(e)
				}
				return true
			}() &&
			func() bool {
				if len(cfg.attemptForError) > 0 {
					_, ok := cfg.attemptForError[e]
					return ok
				}
				return true
			}()
		if !shouldAttempt {
			break
		}

		if cfg.onRetry != nil {
			cfg.onRetry(n, e)
		}
		select {
		case <-time.After(time.Duration(cfg.backoff.Next(n)) * time.Millisecond):
		case <-cfg.ctx.Done():
			errs = append(errs, cfg.ctx.Err())
			break loop
		}
		if len(errs) > 30 {
			errs = errs[:1]
		}
	}

	if cfg.lastErrorOnly {
		return emptyT, errs.Last()
	}
	return emptyT, errs
}
