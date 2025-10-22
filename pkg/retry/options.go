package retry

import (
	"context"
	"time"
)

type retryConfig struct {
	attempts uint
	delay    DelayFunc
	onRetry  func(attempt uint, err error)
	retryIf  func(attempt uint, err error) bool
	context  context.Context
}

func (c *retryConfig) Reset() {
	c.attempts = uint(3)
	c.delay = FixedDelay(time.Millisecond * 100)
	c.onRetry = func(attempt uint, err error) {}
	c.retryIf = nil
	c.context = context.Background()
}

// Option represents an option for retry.
type Option func(*retryConfig)

// Attempts set count of retry. Setting to 0 will retry until the retried function succeeds.
// default is 3
func Attempts(attempts uint) Option {
	return func(c *retryConfig) {
		c.attempts = attempts
	}
}

func Delay(delay DelayFunc) Option {
	return func(c *retryConfig) {
		c.delay = delay
	}
}

func OnRetry(f func(attempt uint, err error)) Option {
	return func(c *retryConfig) {
		if f != nil {
			c.onRetry = f
		}
	}
}

func RetryIf(f func(attempt uint, err error) bool) Option {
	return func(c *retryConfig) {
		if f != nil {
			c.retryIf = f
		}
	}
}

func Context(ctx context.Context) Option {
	return func(c *retryConfig) {
		c.context = ctx
	}
}
