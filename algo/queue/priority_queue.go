package queue

import (
	"container/heap"
	"errors"
	"sync"
)

var ErrEmptyQueue = errors.New("empty queue")

type itemHeap []*item

type item struct {
	value    any
	priority float64
	index    int
}

func (ih *itemHeap) Len() int {
	return len(*ih)
}

func (ih *itemHeap) Less(i, j int) bool {
	return (*ih)[i].priority < (*ih)[j].priority
}

func (ih *itemHeap) Swap(i, j int) {
	(*ih)[i], (*ih)[j] = (*ih)[j], (*ih)[i]
	(*ih)[i].index = i
	(*ih)[j].index = j
}

func (ih *itemHeap) Push(x any) {
	it := x.(*item)
	it.index = len(*ih)
	*ih = append(*ih, it)
}

func (ih *itemHeap) Pop() any {
	old := *ih
	item := old[len(old)-1]
	*ih = old[0 : len(old)-1]
	return item
}

// PriorityQueue represents the queue
type PriorityQueue struct {
	mu       sync.RWMutex
	itemHeap *itemHeap
	lookup   map[any]*item
}

// New initializes an empty priority queue.
func New() PriorityQueue {
	return PriorityQueue{
		itemHeap: &itemHeap{},
		lookup:   make(map[any]*item),
	}
}

// Len returns the number of elements in the queue.
func (pq *PriorityQueue) Len() int {
	return pq.itemHeap.Len()
}

// Insert inserts a new element into the queue. No action is performed on duplicate elements.
func (pq *PriorityQueue) Insert(v any, priority float64) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	_, ok := pq.lookup[v]
	if ok {
		return
	}

	newItem := &item{
		value:    v,
		priority: priority,
	}
	heap.Push(pq.itemHeap, newItem)
	pq.lookup[v] = newItem
}

// Pop removes the element with the highest priority from the queue and returns it.
// In case of an empty queue, an error is returned.
func (pq *PriorityQueue) Pop() (any, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if pq.itemHeap.Len() == 0 {
		return nil, ErrEmptyQueue
	}

	item := heap.Pop(pq.itemHeap).(*item)
	delete(pq.lookup, item.value)
	return item.value, nil
}

// UpdatePriority changes the priority of a given item.
// If the specified item is not present in the queue, no action is performed.
func (pq *PriorityQueue) UpdatePriority(x any, newPriority float64) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	item, ok := pq.lookup[x]
	if !ok {
		return
	}

	item.priority = newPriority
	heap.Fix(pq.itemHeap, item.index)
}
