package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDelay(t *testing.T) {
	for _, v := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"fixed",
			func(t *testing.T) {
				for i := 0; i < 100; i++ {
					assert.Equal(t, time.Second, FixedDelay(time.Second)(uint(i)))
				}
			},
		},
		{
			"random",
			func(t *testing.T) {
				for i := 0; i < 10; i++ {
					assert.Greater(t, time.Second, RandomDelay(time.Second)(uint(i)))
				}
			},
		},
		{
			"backoff",
			func(t *testing.T) {
				arr := []time.Duration{time.Second, time.Second << 1, time.Second << 2, time.Second << 3, time.Second << 3, time.Second << 3}
				for i := 0; i < len(arr); i++ {
					assert.Equal(t, arr[i], BackOffDelay(time.Second, 3)(uint(i)))
				}
			},
		},
	} {
		t.Run(v.name, v.f)
	}
}

func TestRetry(t *testing.T) {
	err := errors.New("fake error")

	for _, v := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"retry_default_3",
			func(t *testing.T) {
				i := 0
				Do(func() error { i++; return err })
				assert.Equal(t, 3, i)
			},
		},
		{
			"retry_attempt_5",
			func(t *testing.T) {
				i := 0
				Do(func() error { i++; return err }, Attempts(5))
				assert.Equal(t, 5, i)
			},
		},
		{
			"retry_context",
			func(t *testing.T) {
				i := 0
				ctx, cancel := context.WithCancel(context.Background())
				Do(func() error {
					i++
					if i > 2 {
						cancel()
					}
					return err
				}, Context(ctx))
				assert.Equal(t, 3, i)
			},
		},
		{
			"retry_if",
			func(t *testing.T) {
				i := 0
				Do(func() error {
					i++
					return err
				}, Attempts(0), RetryIf(func(attempt uint, err error) bool { return attempt < 3 }))
				assert.Equal(t, 3, i)
			},
		},
	} {
		t.Run(v.name, v.f)
	}
}
