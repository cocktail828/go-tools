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
	if len(b.nodes) == 0 {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	pos := b.pos.Load()
	for i := range uint32(len(b.nodes)) {
		if n := b.nodes[(i+pos)%uint32(len(b.nodes))]; n.Healthy() {
			return WrapNode{Node: n, nodeArrayRemove: b}
		}
	}

	return nil
}
