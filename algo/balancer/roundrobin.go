package balancer

import (
	"sync"
)

type rrBalancer struct {
	mu    sync.RWMutex
	pos   uint16
	array []Node
}

func NewRR(array []Node) Balancer {
	return &rrBalancer{array: array}
}

func (b *rrBalancer) Pick() (n Node) {
	b.mu.RLock()
	array := b.array
	b.mu.RUnlock()

	if len(array) == 0 {
		return nil
	}
	c := array[b.pos%uint16(len(array))]
	b.pos++
	return c
}
