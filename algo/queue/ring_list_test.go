package queue

import (
	"strconv"
	"sync"
	"testing"

	"github.com/cocktail828/go-tools/z/mathx"
	"github.com/stretchr/testify/assert"
)

func TestBasicOperationsRingQueue(t *testing.T) {
	rq := NewRingQueue(4)
	assert.NotNil(t, rq, "Queue should be created successfully")
	assert.Equal(t, 0, rq.Len(), "New queue should be empty")
	assert.True(t, rq.IsEmpty(), "New queue should be empty")
	assert.False(t, rq.IsFull(), "New queue should not be full")

	// Test Push
	assert.True(t, rq.Push(1), "Should be able to push first element")
	assert.Equal(t, 1, rq.Len(), "Length should be 1 after pushing one element")
	assert.False(t, rq.IsEmpty(), "Queue should not be empty after pushing")

	// Test Peek
	assert.Equal(t, 1, rq.Peek(), "Peek should return the first element")
	assert.Equal(t, 1, rq.Len(), "Length should remain the same after Peek")

	// Test Poll
	assert.Equal(t, 1, rq.Poll(), "Poll should return the first element")
	assert.Equal(t, 0, rq.Len(), "Length should be 0 after polling the only element")
	assert.True(t, rq.IsEmpty(), "Queue should be empty after polling all elements")
}

// TestNilValue handles nil value testing
func TestNilValue(t *testing.T) {
	rq := NewRingQueue(4)
	assert.False(t, rq.Push(nil), "Push should return false for nil value")
	assert.Equal(t, 0, rq.Len(), "Length should remain 0 after trying to push nil")
}

// TestCircularBehavior tests the circular behavior of the queue
func TestCircularBehavior(t *testing.T) {
	capacity := 4
	rq := NewRingQueue(capacity)

	// Fill the queue
	for i := 1; i <= capacity; i++ {
		assert.True(t, rq.Push(i), "Should be able to push element")
	}
	assert.Equal(t, capacity, rq.Len(), "Queue should be full")
	assert.True(t, rq.IsFull(), "IsFull should return true when queue is full")

	// Test overwrite behavior
	for i := capacity + 1; i <= capacity*2; i++ {
		assert.True(t, rq.Push(i), "Should be able to push to full queue (overwrite)")
	}
	assert.Equal(t, capacity, rq.Len(), "Length should remain the same after overwriting")

	// Verify the queue now contains the last 'capacity' elements
	expected := []int{5, 6, 7, 8}
	for i := 0; i < capacity; i++ {
		val := rq.Poll().(int)
		assert.Equal(t, expected[i], val, "Poll should return overwritten elements in order")
	}
}

// TestEdgeCases tests edge cases for the queue
func TestEdgeCases(t *testing.T) {
	// Test with zero capacity
	rq := NewRingQueue(0)
	assert.Nil(t, rq.Poll(), "Poll on empty queue should return nil")

	// Test with negative capacity
	rq = NewRingQueue(-1)
	assert.Nil(t, rq.Poll(), "Poll on queue with negative capacity should return nil")

	// Test Poll on empty queue
	rq = NewRingQueue(4)
	assert.Nil(t, rq.Poll(), "Poll on empty queue should return nil")

	// Test Peek on empty queue
	assert.Nil(t, rq.Peek(), "Peek on empty queue should return nil")

	// Test Clear on empty queue
	rq.Clear()
	assert.True(t, rq.IsEmpty(), "Clear on empty queue should work")

	// Test Clear on non-empty queue
	rq.Push(1)
	rq.Push(2)
	rq.Clear()
	assert.True(t, rq.IsEmpty(), "Queue should be empty after Clear")
	assert.Equal(t, 0, rq.Len(), "Length should be 0 after Clear")
}

// TestConcurrency tests the queue under concurrent access
func TestConcurrency(t *testing.T) {
	capacity := 100
	rq := NewRingQueue(capacity)
	var wg sync.WaitGroup
	operationCount := 1000

	// Producer goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for j := 0; j < operationCount; j++ {
				value := strconv.Itoa(producerID*10000 + j)
				rq.Push(value)
			}
		}(i)
	}

	// Consumer goroutines
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(consumerID int) {
			defer wg.Done()
			for j := 0; j < operationCount; j++ {
				rq.Poll()
			}
		}(i)
	}

	// Wait for all operations to complete
	wg.Wait()

	// Verify the queue is not in an inconsistent state
	length := rq.Len()
	assert.GreaterOrEqual(t, length, 0, "Queue length should be non-negative")
	assert.LessOrEqual(t, length, int(mathx.Next2Power(int64(capacity))), "Queue length should not exceed capacity")

	// Clean up remaining elements to check for consistency
	elements := make(map[any]bool)
	for !rq.IsEmpty() {
		val := rq.Poll()
		elements[val] = true
	}
	assert.True(t, rq.IsEmpty(), "Queue should be empty after polling all elements")
}

// TestRingListOriginal maintains the original test logic
func TestRingListOriginal(t *testing.T) {
	cnt := 4
	rq := NewRingQueue(cnt)
	nodes := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i, n := range nodes {
		if i < cnt {
			assert.True(t, rq.Push(n))
		} else {
			assert.True(t, rq.Push(n), "With the new implementation, Push should always return true")
		}
	}

	// test try push
	for range 4 {
		assert.False(t, rq.TryPush(10))
	}

	// Since we now overwrite, we'll get the last 4 elements (6,7,8,9) in order
	expected := []int{6, 7, 8, 9}
	for i := range 4 {
		assert.Equal(t, expected[i], rq.Poll().(int))
	}

	for range 10 {
		rq.Poll()
	}

	for range 10 {
		assert.Nil(t, rq.Poll())
	}
}
