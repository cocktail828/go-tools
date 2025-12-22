package queue

import (
	"sync"

	"github.com/cocktail828/go-tools/algo/mathx"
)

// RingQueue implements a ring queue using a fixed-size array,
// and supports concurrent access.
// When queue is full, new elements will overwrite the oldest ones
type RingQueue struct {
	mu     sync.RWMutex // Read-write lock for concurrency safety
	buffer []any        // Buffer for storing elements
	max    int          // Maximum capacity
	size   int          // Current number of elements
	head   int          // Head index
	tail   int          // Tail index
}

// NewRingQueue creates and initializes a circular queue with max capacity
// max will be round up to power of 2
func NewRingQueue(max int) *RingQueue {
	if max <= 0 {
		max = 1024
	}

	max = int(mathx.Next2Power(int64(max)))
	return &RingQueue{
		buffer: make([]any, max),
		max:    max,
		head:   0,
		tail:   0,
		size:   0,
	}
}

// TryPush attempts to add element n to the end of the queue
// If queue is full, it will return directly
// Returns true if addition is successful
func (rq *RingQueue) TryPush(n any) bool {
	if n == nil {
		return false
	}

	// return fastly on queue full
	if rq.IsFull() {
		return false
	}

	rq.mu.Lock()
	defer rq.mu.Unlock()

	// If queue is full and new element will not be added
	if rq.size == rq.max {
		return false
	}

	// Place new element at tail
	rq.buffer[rq.tail] = n
	// Update tail pointer
	rq.tail = (rq.tail + 1) % rq.max
	rq.size++

	return true
}

// Push adds element n to the end of the queue
// If queue is full, the oldest element will be overwritten
// Returns false if n is nil
// Returns true if addition is successful
func (rq *RingQueue) Push(n any) bool {
	if n == nil {
		return false
	}

	rq.mu.Lock()
	defer rq.mu.Unlock()

	// Place new element at tail
	rq.buffer[rq.tail] = n
	// Update tail pointer
	rq.tail = (rq.tail + 1) % rq.max

	// If queue is full and a new element is added, move head pointer
	if rq.size == rq.max {
		rq.head = (rq.head + 1) % rq.max
	} else {
		// Queue is not full, increase element count
		rq.size++
	}

	return true
}

// Poll removes and returns an element from the head of the queue
// Returns nil if queue is empty
func (rq *RingQueue) Poll() any {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if rq.size == 0 {
		return nil
	}

	// Get head element
	val := rq.buffer[rq.head]
	// Clear value to help garbage collection
	rq.buffer[rq.head] = nil
	// Update head pointer
	rq.head = (rq.head + 1) % rq.max
	// Decrease element count
	rq.size--

	return val
}

// Len returns the current number of elements in the queue
func (rq *RingQueue) Len() int {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return rq.size
}

// IsEmpty checks if the queue is empty
func (rq *RingQueue) IsEmpty() bool {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return rq.size == 0
}

// IsFull checks if the queue is full
func (rq *RingQueue) IsFull() bool {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return rq.size == rq.max
}

// Clear removes all elements from the queue
func (rq *RingQueue) Clear() {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	// Clear all elements to help garbage collection
	for i := 0; i < rq.size; i++ {
		idx := (rq.head + i) % rq.max
		rq.buffer[idx] = nil
	}

	// Reset all pointers and counters
	rq.head = 0
	rq.tail = 0
	rq.size = 0
}

// Peek returns the head element without removing it
// Returns nil if queue is empty
func (rq *RingQueue) Peek() any {
	rq.mu.RLock()
	defer rq.mu.RUnlock()

	if rq.size == 0 {
		return nil
	}

	return rq.buffer[rq.head]
}
