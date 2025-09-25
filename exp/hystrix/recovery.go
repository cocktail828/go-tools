package hystrix

import (
	"slices"
	"sync"
)

type recovery struct {
	mu    sync.RWMutex
	next  int
	array []bool
}

func (r *recovery) Update(ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.array[r.next] = ok
	r.next = (r.next + 1) % len(r.array)
}

func (r *recovery) IsHealthy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return !slices.Contains(r.array, false)
}

func (r *recovery) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.next = 0
	r.array = make([]bool, len(r.array))
}
