package queue

import (
	"container/list"
	"sync"
)

type RingQueue struct {
	mu   sync.RWMutex
	max  int           // 最大队列长度
	list *list.List    // 队列头节点（用于遍历和删除）
	curr *list.Element // 当前访问节点（用于 Poll）
}

func NewRingQueue(max int) *RingQueue {
	return &RingQueue{
		max:  max,
		list: list.New(),
	}
}

func (rq *RingQueue) Insert(n any) bool {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if n == nil || rq.list.Len() >= rq.max {
		return false
	}

	rq.list.PushBack(n)
	if rq.curr == nil {
		rq.curr = rq.list.Front()
	}

	return true
}

func (rq *RingQueue) Remove(n *list.Element) {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if n == nil || rq.list == nil {
		return
	}

	if n == rq.curr {
		if length := rq.list.Len(); length > 1 {
			rq.pollLocked()
		} else {
			rq.curr = nil
		}
	}
	rq.list.Remove(n)
}

func (rq *RingQueue) Poll() *list.Element {
	rq.mu.Lock()
	defer rq.mu.Unlock()
	return rq.pollLocked()
}

func (rq *RingQueue) pollLocked() *list.Element {
	if rq.curr == nil {
		return nil
	}

	node := rq.curr
	if next := node.Next(); next != nil {
		rq.curr = node.Next()
	} else {
		rq.curr = rq.list.Front()
	}
	return node
}

func (rq *RingQueue) Len() int {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return rq.list.Len()
}

func (rq *RingQueue) IsEmpty() bool {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return rq.list.Len() == 0
}

func (rq *RingQueue) IsFull() bool {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return rq.list.Len() >= rq.max
}

func (rq *RingQueue) Clear() {
	rq.mu.Lock()
	defer rq.mu.Unlock()
	rq.list.Init()
	rq.curr = nil
}
