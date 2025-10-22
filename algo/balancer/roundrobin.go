package balancer

type rrBalancer struct {
	nodeArray
	pos uint32
}

func NewRR(nodes []Node) Balancer {
	if nodes == nil {
		nodes = []Node{}
	}
	return &rrBalancer{nodeArray: nodeArray{nodes: nodes}}
}

func (b *rrBalancer) Pick() Node {
	if len(b.nodes) == 0 {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for i := range uint32(len(b.nodes)) {
		if n := b.nodes[(i+b.pos)%uint32(len(b.nodes))]; n.Healthy() {
			b.pos += i + 1
			return WrapNode{Node: n, nodeArrayRemove: b}
		}
	}

	return nil
}
