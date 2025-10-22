package retry

import (
	"math/rand"
	"sync"
	"time"
)

type DelayFunc func(attempt uint) time.Duration

var (
	randMu  sync.Mutex
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func BackoffDelay(initialDelay time.Duration, maxDelay time.Duration) DelayFunc {
	if initialDelay <= 0 {
		initialDelay = time.Millisecond * 100
	}
	if maxDelay <= 0 {
		maxDelay = time.Second * 10
	}

	return func(attempt uint) time.Duration {
		backoff := initialDelay * (1 << min(attempt, uint(30)))

		randMu.Lock()
		jitter := time.Duration(randGen.Int63n(int64(backoff)))
		randMu.Unlock()

		// random jitter (backoff * (0.9 ~ 1.1))
		backoff += jitter/5 - backoff/10
		return max(backoff, maxDelay)
	}
}

func FixedDelay(v time.Duration) DelayFunc {
	if v <= 0 {
		v = time.Millisecond * 100
	}
	return func(_ uint) time.Duration { return v }
}

func RandomDelay(maxDelay time.Duration) DelayFunc {
	if maxDelay <= 0 {
		maxDelay = time.Second * 10
	}
	return func(_ uint) time.Duration {
		randMu.Lock()
		defer randMu.Unlock()
		return time.Duration(randGen.Int63n(int64(maxDelay)))
	}
}
