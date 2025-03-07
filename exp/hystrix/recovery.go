package hystrix

import (
	"slices"
	"sync"
)

type Recovery struct {
	mu    sync.RWMutex
	next  int
	array []bool
}

func NewRecovery(maxprobe int) *Recovery {
	return &Recovery{
		array: make([]bool, maxprobe),
	}
}

func (r *Recovery) Update(healthy bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.array[r.next] = healthy
	r.next = (r.next + 1) % len(r.array)
}

func (r *Recovery) IsHealthy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return !slices.Contains(r.array, false)
}

func (r *Recovery) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.next = 0
	r.array = make([]bool, len(r.array))
}
