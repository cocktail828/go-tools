package retry

import (
	"math"
	"math/rand"
	"time"
)

type DelayFunc func(attempt uint) time.Duration

func init() {
	rand.Seed(time.Now().UnixNano())
}

func BackOffDelay(delay time.Duration, maxBackOffN uint) DelayFunc {
	if delay <= 0 {
		delay = time.Millisecond * 100
	}

	maxPossibleAttempt := uint(math.Log2(float64(math.MaxInt64 / int64(delay))))
	if maxBackOffN == 0 || maxBackOffN > maxPossibleAttempt {
		maxBackOffN = maxPossibleAttempt
	}

	return func(attempt uint) time.Duration {
		return delay * (1 << min(attempt, maxBackOffN))
	}
}

func FixedDelay(v time.Duration) DelayFunc {
	if v <= 0 {
		v = time.Millisecond * 100
	}
	return func(attempt uint) time.Duration { return v }
}

func RandomDelay(maxJitter time.Duration) DelayFunc {
	return func(attempt uint) time.Duration {
		return time.Duration(rand.Int63n(int64(maxJitter)))
	}
}

func min(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}
