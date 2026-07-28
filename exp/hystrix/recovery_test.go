package hystrix

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecovery_DefaultCapacity(t *testing.T) {
	// non-positive capacity falls back to the default (5)
	r := newRecovery(0)
	assert.Equal(t, 5, r.capacity)

	r = newRecovery(-3)
	assert.Equal(t, 5, r.capacity)
}

func TestRecovery_NotHealthyUntilFull(t *testing.T) {
	r := newRecovery(4)

	// buffer not yet full => never healthy, regardless of success
	for i := 0; i < 3; i++ {
		r.Update(true)
		assert.False(t, r.IsHealthy(), "should not be healthy before buffer is full")
	}

	// fourth success fills the buffer at 100% => healthy
	r.Update(true)
	assert.True(t, r.IsHealthy())
}

func TestRecovery_HealthyThreshold(t *testing.T) {
	r := newRecovery(5)

	// 4 success, 1 fail => 80% == threshold => healthy
	r.Update(true)
	r.Update(true)
	r.Update(true)
	r.Update(true)
	r.Update(false)
	assert.True(t, r.IsHealthy())
}

func TestRecovery_BelowThreshold(t *testing.T) {
	r := newRecovery(5)

	// 3 success, 2 fail => 60% < 80% => not healthy
	r.Update(true)
	r.Update(true)
	r.Update(true)
	r.Update(false)
	r.Update(false)
	assert.False(t, r.IsHealthy())
}

func TestRecovery_RingEviction(t *testing.T) {
	r := newRecovery(3)

	// fill with failures => unhealthy
	r.Update(false)
	r.Update(false)
	r.Update(false)
	assert.False(t, r.IsHealthy())

	// overwrite the oldest entries with successes; ring keeps only last 3
	r.Update(true)
	r.Update(true)
	r.Update(true)
	assert.True(t, r.IsHealthy())
	assert.EqualValues(t, 3, r.successCount)
	assert.EqualValues(t, 3, r.size)
}

func TestRecovery_Reset(t *testing.T) {
	r := newRecovery(3)
	r.Update(true)
	r.Update(true)
	r.Update(true)
	assert.True(t, r.IsHealthy())

	r.Reset()
	assert.EqualValues(t, 0, r.size)
	assert.EqualValues(t, 0, r.successCount)
	assert.EqualValues(t, 0, r.writeIndex)
	assert.False(t, r.IsHealthy())
}

func TestRecovery_String(t *testing.T) {
	r := newRecovery(4)
	r.Update(true)
	r.Update(false)

	s := r.String()
	assert.True(t, strings.Contains(s, "success: 1"))
	assert.True(t, strings.Contains(s, "total: 2"))
	assert.True(t, strings.Contains(s, "capacity: 4"))
}

func TestRecovery_ConcurrentUpdate(t *testing.T) {
	r := newRecovery(16)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				r.Update(n%2 == 0)
				r.IsHealthy()
				_ = r.String()
			}
		}(i)
	}
	wg.Wait()

	// invariant: successCount never exceeds size, size never exceeds capacity
	r.mu.RLock()
	defer r.mu.RUnlock()
	assert.LessOrEqual(t, r.successCount, r.size)
	assert.LessOrEqual(t, r.size, r.capacity)
}
