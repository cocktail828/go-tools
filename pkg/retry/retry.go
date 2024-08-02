package retry

import (
	"context"
	"time"
)

func Do(f func() error, opts ...Option) error {
	_, err := DoWithData(func() (any, error) {
		return nil, f()
	}, opts...)
	return err
}

func DoWithData[T any](f func() (T, error), opts ...Option) (T, error) {
	cfg := &retryConfig{
		attempts: uint(3),
		delay:    FixedDelay(time.Millisecond * 10),
		onRetry:  func(attempt uint, err error) {},
		retryIf:  nil,
		context:  context.Background(),
	}

	for _, f := range opts {
		f(cfg)
	}

	var n uint = 0
	errs := Error{}
	for {
		t, err := f()
		if err == nil {
			return t, nil
		}
		n++
		errs = append(errs, err)

		if cfg.attempts > 0 && n >= cfg.attempts {
			return t, errs
		}

		if cfg.retryIf != nil && !cfg.retryIf(n, err) {
			return t, err
		}

		cfg.onRetry(n, err)
		select {
		case <-time.After(cfg.delay(n, err)):
		case <-cfg.context.Done():
			return t, errs
		}
	}
}
