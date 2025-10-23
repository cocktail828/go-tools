package balancer

type failoverBalancer struct {
	*nodeArray
	pos uint32
}

func NewFailover(nodes []Node) Balancer {
	return &failoverBalancer{nodeArray: &nodeArray{nodes: nodes}}
}

func (b *failoverBalancer) String() string {
	return "failover"
}

func (b *failoverBalancer) Pick() Node {
	if len(b.nodes) == 0 {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for i := range uint32(len(b.nodes)) {
		idx := (i + b.pos) % uint32(len(b.nodes))
		if n := b.nodes[idx]; n.Healthy() {
			b.pos = idx
			return fallibleNode{Node: n, nodeArray: b.nodeArray}
		}
	}

	return nil
}
