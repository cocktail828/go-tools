package retry

import (
	"math"
	"math/rand"
	"time"
)

type DelayFunc func(attempt uint, err error) time.Duration

// BackOffDelay is a DelayType which increases delay between consecutive retries
// Delay set delay between retry, default is 100ms
func BackOffDelay(delay time.Duration, maxBackOffN uint) DelayFunc {
	// 1 << 63 would overflow signed int64 (time.Duration), thus 62.
	const max uint = 62
	if delay <= 0 {
		delay = 1
	}

	if maxBackOffN == 0 {
		maxBackOffN = max - uint(math.Floor(math.Log2(float64(delay))))
	}

	return func(attempt uint, err error) time.Duration {
		if attempt > maxBackOffN {
			attempt = maxBackOffN
		}
		return delay << attempt
	}
}

// FixedDelay is a DelayType which keeps delay the same through all iterations
func FixedDelay(v time.Duration) DelayFunc {
	return func(attempt uint, err error) time.Duration { return v }
}

// RandomDelay is a DelayType which picks a random delay up to maxJitter
func RandomDelay(maxJitter time.Duration) DelayFunc {
	return func(attempt uint, err error) time.Duration {
		return time.Duration(rand.Int63n(int64(maxJitter)))
	}
}
