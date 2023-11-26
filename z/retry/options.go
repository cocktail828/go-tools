package retry

import (
	"context"

	"github.com/cocktail828/go-tools/z/retry/backoff"
)

type config struct {
	attempts        int
	attemptForError map[error]struct{}
	onRetry         func(n int, err error) // Function signature of OnRetry function
	retryIf         func(error) bool       // Function signature of retry if function
	// return the next delay to wait after the retriable function fails on `err` after `n` attempts.
	backoff       backoff.BackOff
	lastErrorOnly bool
	ctx           context.Context
}

// Option represents an option for retry.
type Option func(*config)

// return the direct last error that came from the retried function
// default is false (return wrapped errors with everything)
func LastErrorOnly(lastErrorOnly bool) Option {
	return func(c *config) {
		c.lastErrorOnly = lastErrorOnly
	}
}

// Attempts set count of retry. 0 for infinity loop until success.
func Attempts(attempts int) Option {
	return func(c *config) {
		c.attempts = attempts
	}
}

// AttemptForError sets count of retry in case execution results in given `err`
// Retries for the given `err` are also counted against total retries.
// The retry will stop if any of given retries is exhausted.
func AttemptForError(err error) Option {
	return func(c *config) {
		c.attemptForError[err] = struct{}{}
	}
}

// OnRetry function callback are called each retry
func OnRetry(f func(n int, err error)) Option {
	return func(c *config) {
		if f != nil {
			c.onRetry = f
		}
	}
}

// RetryIf controls whether a retry should be attempted after an error
// (assuming there are any retry attempts remaining)
func RetryIf(f func(error) bool) Option {
	return func(c *config) {
		if f != nil {
			c.retryIf = f
		}
	}
}

// Context allow to set context of retry
// default are Background context
// example of immediately cancellation (maybe it isn't the best example, but it describes behavior enough; I hope)
func Context(ctx context.Context) Option {
	return func(c *config) {
		c.ctx = ctx
	}
}

func BackOff(b backoff.BackOff) Option {
	return func(c *config) {
		c.backoff = b
	}
}
