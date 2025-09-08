package balancer

import (
	"sync"
	"sync/atomic"
)

type failoverBalancer struct {
	mu    sync.RWMutex
	pos   atomic.Uint32
	array []Node
}

func NewFailover(array []Node) Balancer {
	return &failoverBalancer{array: array}
}

func (b *failoverBalancer) Pick() (n Node) {
	b.mu.RLock()
	array := b.array
	b.mu.RUnlock()

	for i := b.pos.Load(); i < uint32(len(array)); i++ {
		n := array[i]
		if h, ok := n.(Healthy); ok && h.Healthy() {
			b.pos.Store(i)
			return n
		}
	}
	return nil
}
