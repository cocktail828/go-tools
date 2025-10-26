package retry

import (
	"fmt"
	"sync"
	"time"
)

var configPool = sync.Pool{
	New: func() any {
		return &retryConfig{}
	},
}

func Do(f func() error, opts ...Option) error {
	_, err := DoWithData(func() (any, error) {
		return nil, f()
	}, opts...)
	return err
}

func DoWithData[T any](f func() (T, error), opts ...Option) (T, error) {
	cfg := configPool.Get().(*retryConfig)
	defer configPool.Put(cfg)

	cfg.Reset()
	for _, opt := range opts {
		opt(cfg)
	}

	return retry(f, cfg)
}

func retry[T any](f func() (T, error), cfg *retryConfig) (t T, oerr error) {
	var n uint = 0
	errs := Error{}

	defer func() {
		if err := recover(); err != nil {
			errs = append(errs, fmt.Errorf("%v", err))
			oerr = errs
		}
	}()

	for {
		t, err := f()
		if err == nil {
			return t, nil
		}
		n++
		errs = append(errs, err)

		// if attempts is 0, retry forever
		// if attempts is not 0, retry until attempts
		if cfg.attempts > 0 && n >= cfg.attempts {
			return t, errs
		}

		// if retryIf is nil, retry always
		// if retryIf is not nil, retry if retryIf returns true
		if cfg.retryIf != nil && !cfg.retryIf(n, err) {
			return t, err
		}

		// if context is nil, retry until delay
		// if context is not nil, retry until context is done or delay
		select {
		case <-time.After(cfg.delay(n)):
		case <-cfg.context.Done():
			return t, errs
		}
		cfg.onRetry(n, err)
	}
}
