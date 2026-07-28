package queue

import (
	"container/heap"
	"errors"
	"sync"
)

var ErrEmptyQueue = errors.New("empty priority queue")

type itemHeap[T comparable] []*item[T]

type item[T comparable] struct {
	value    T
	priority float64
	index    int
}

func (ih *itemHeap[T]) Len() int {
	return len(*ih)
}

func (ih *itemHeap[T]) Less(i, j int) bool {
	return (*ih)[i].priority < (*ih)[j].priority
}

func (ih *itemHeap[T]) Swap(i, j int) {
	(*ih)[i], (*ih)[j] = (*ih)[j], (*ih)[i]
	(*ih)[i].index = i
	(*ih)[j].index = j
}

func (ih *itemHeap[T]) Push(x any) {
	it := x.(*item[T])
	it.index = len(*ih)
	*ih = append(*ih, it)
}

func (ih *itemHeap[T]) Pop() any {
	old := *ih
	n := len(old)
	it := old[n-1]
	old[n-1] = nil // avoid memory leak
	*ih = old[0 : n-1]
	return it
}

// PriorityQueue represents the queue
type PriorityQueue[T comparable] struct {
	mu       sync.RWMutex
	itemHeap *itemHeap[T]
	lookup   map[T]*item[T]
}

// NewPriorityQueue initializes an empty priority queue.
// The lower the priority value, the higher the priority.
func NewPriorityQueue[T comparable]() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		itemHeap: &itemHeap[T]{},
		lookup:   make(map[T]*item[T]),
	}
}

// Len returns the number of elements in the queue.
func (pq *PriorityQueue[T]) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return pq.itemHeap.Len()
}

// Clear the queue, removing all elements.
func (pq *PriorityQueue[T]) Clear() {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.itemHeap = &itemHeap[T]{}
	pq.lookup = make(map[T]*item[T])
}

// TryPush attempts to insert an element with the given priority.
// If the element already exists in the queue, it returns false.
func (pq *PriorityQueue[T]) TryPush(v T, priority float64) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	_, ok := pq.lookup[v]
	if ok {
		return false
	}

	newItem := &item[T]{
		value:    v,
		priority: priority,
	}
	heap.Push(pq.itemHeap, newItem)
	pq.lookup[v] = newItem
	return true
}

// Push inserts a new element into the queue. No action is performed on duplicate elements.
func (pq *PriorityQueue[T]) Push(v T, priority float64) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	_, ok := pq.lookup[v]
	if ok {
		return
	}

	newItem := &item[T]{
		value:    v,
		priority: priority,
	}
	heap.Push(pq.itemHeap, newItem)
	pq.lookup[v] = newItem
}

// Pop removes the element with the highest priority from the queue and returns it.
// In case of an empty queue, an error is returned.
func (pq *PriorityQueue[T]) Pop() (T, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if pq.itemHeap.Len() == 0 {
		var zero T
		return zero, ErrEmptyQueue
	}

	it := heap.Pop(pq.itemHeap).(*item[T])
	delete(pq.lookup, it.value)
	return it.value, nil
}

// UpdatePriority changes the priority of a given item.
// If the specified item is not present in the queue, no action is performed.
func (pq *PriorityQueue[T]) UpdatePriority(x T, newPriority float64) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	it, ok := pq.lookup[x]
	if !ok {
		return
	}

	it.priority = newPriority
	heap.Fix(pq.itemHeap, it.index)
}
