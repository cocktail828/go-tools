package balancer

import "sync/atomic"

type rrBalancer struct {
	nodeArray
	pos atomic.Uint32
}

func NewRR(nodes []Node) Balancer {
	if nodes == nil {
		nodes = []Node{}
	}
	return &rrBalancer{nodeArray: nodeArray{nodes: nodes}}
}

func (b *rrBalancer) Pick() Node {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.Empty() {
		return nil
	}

	pos := b.pos.Add(1) % uint32(b.Len())
	return b.nodes[pos]
}
