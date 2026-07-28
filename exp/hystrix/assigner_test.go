package hystrix

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssigner_AcquireRelease(t *testing.T) {
	a := &Assigner{maxCount: 2}

	assert.True(t, a.TryAcquire())
	assert.True(t, a.TryAcquire())
	assert.EqualValues(t, 2, a.Allocated())
	assert.EqualValues(t, 0, a.Available())

	// over capacity
	assert.False(t, a.TryAcquire())
	assert.EqualValues(t, 2, a.Allocated())

	a.Release()
	assert.EqualValues(t, 1, a.Allocated())
	assert.EqualValues(t, 1, a.Available())
	assert.True(t, a.TryAcquire())
}

func TestAssigner_ReleaseFloor(t *testing.T) {
	a := &Assigner{maxCount: 2}
	// Release without acquire must not underflow below zero
	a.Release()
	a.Release()
	assert.EqualValues(t, 0, a.Allocated())
}

func TestAssigner_Resize(t *testing.T) {
	a := &Assigner{maxCount: 1}
	assert.True(t, a.TryAcquire())
	assert.False(t, a.TryAcquire())

	// grow capacity, more tickets become available
	a.Resize(3)
	assert.True(t, a.TryAcquire())
	assert.True(t, a.TryAcquire())
	assert.False(t, a.TryAcquire())

	// shrink below allocated: Available may be negative, no new tickets granted
	a.Resize(1)
	assert.EqualValues(t, 3, a.Allocated())
	assert.EqualValues(t, -2, a.Available())
	assert.False(t, a.TryAcquire())
}

func TestAssigner_ResizeNegativePanics(t *testing.T) {
	a := &Assigner{maxCount: 1}
	assert.Panics(t, func() { a.Resize(-1) })
}

func TestAssigner_ConcurrentAcquireRelease(t *testing.T) {
	const max = 8
	a := &Assigner{maxCount: max}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if a.TryAcquire() {
					a.Release()
				}
			}
		}()
	}
	wg.Wait()

	// all tickets returned, invariant holds
	assert.EqualValues(t, 0, a.Allocated())
	assert.EqualValues(t, max, a.Available())
}
