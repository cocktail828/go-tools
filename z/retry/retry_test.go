package retry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z/retry"
	"github.com/cocktail828/go-tools/z/retry/backoff"
)

func TestRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fmt.Println(retry.Do(func(ctx context.Context) error {
		return errors.New("xxx")
		// return nil
	}, retry.Attempts(3),
		retry.BackOff(&backoff.Exponential{}),
		retry.OnRetry(func(n int, err error) {}),
		retry.RetryIf(func(err error) bool { return true }),
		retry.Context(ctx),
		retry.LastErrorOnly(true),
	))
}
