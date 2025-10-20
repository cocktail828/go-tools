package balancer

import (
	"sync/atomic"
)

type failoverBalancer struct {
	nodeArray
	pos atomic.Uint32
}

func NewFailover(nodes []Node) Balancer {
	return &failoverBalancer{nodeArray: nodeArray{nodes: nodes}}
}

func (b *failoverBalancer) Pick() Node {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.Empty() {
		return nil
	}

	pos := b.pos.Load()
	for i := range uint32(b.Len()) {
		n := b.nodes[(i+pos)%uint32(b.Len())]
		h, ok := n.(Healthy)
		if !ok {
			return n
		}

		if h.Healthy() {
			b.pos.Store(i)
			return n
		}
	}

	return nil
}
