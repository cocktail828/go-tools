package hystrix

import (
	"fmt"
	"sync"
)

// recovery implment a recovery status manager for hystrix circuitbreaker
// It use a ring buffer to store the recent probe results.
// The recovery status is based on the recent probe results.
// If the success rate of the recent probes is higher than 80%,
// the circuitbreaker will be considered as healthy.
type recovery struct {
	mu           sync.RWMutex
	buffer       []bool
	capacity     int
	size         int
	writeIndex   int
	successCount int
}

func newRecovery(capacity int) *recovery {
	if capacity <= 0 {
		capacity = 5 // default capacity
	}
	return &recovery{
		buffer:   make([]bool, capacity),
		capacity: capacity,
	}
}

// String show the recovery status, including the following fields:
// successCount: 成功的探测次数
// total: 总探测次数
// success_rate: 成功率（%）
// capacity: 缓冲区容量
func (r *recovery) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	successRate := 0.0
	if r.size > 0 {
		successRate = float64(r.successCount) * 100.0 / float64(r.size)
	}

	return fmt.Sprintf("{success: %d, total: %d, success_rate: %.1f%%, capacity: %d}",
		r.successCount, r.size, successRate, r.capacity)
}

// Update update the recovery status with a new probe result
// ok indicate whether the probe is successful
func (r *recovery) Update(ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.size == r.capacity {
		// decrease success count if the old data is successful
		if r.buffer[r.writeIndex] {
			r.successCount--
		}
	} else {
		r.size++
	}

	r.buffer[r.writeIndex] = ok
	if ok {
		r.successCount++
	}

	r.writeIndex = (r.writeIndex + 1) % r.capacity
}

// IsHealthy check whether the system is healthy based on the recent probe results.
// The system is considered as healthy if the success rate of the recent probes is higher than 80%.
func (r *recovery) IsHealthy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// If the size of the buffer is less than the capacity,
	// In this case, the system is considered as not healthy.
	if r.size < r.capacity {
		return false
	}

	successRate := float64(r.successCount) / float64(r.size)
	return successRate >= 0.8
}

// Reset reset the recovery status to the initial state.
func (r *recovery) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size = 0
	r.writeIndex = 0
	r.successCount = 0
}
