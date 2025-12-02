package balancer

type rrBalancer struct {
	Candidate
	pos uint32
}

func NewRoundRobin(nodes []Node) Balancer {
	if nodes == nil {
		nodes = []Node{}
	}
	return &rrBalancer{Candidate: Candidate{nodes: nodes}}
}

func (b *rrBalancer) String() string {
	return "roundrobin"
}

func (b *rrBalancer) Pick() Node {
	if len(b.nodes) == 0 {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for i := range uint32(len(b.nodes)) {
		idx := (i + b.pos) % uint32(len(b.nodes))
		if n := b.nodes[idx]; n.Healthy() {
			b.pos = idx + 1
			return fallibleNode{n, &b.Candidate}
		}
	}

	return nil
}
