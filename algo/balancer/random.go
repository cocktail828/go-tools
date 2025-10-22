package balancer

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type randomBalancer struct {
	nodeArray
}

func NewRandom(nodes []Node) Balancer {
	return &randomBalancer{nodeArray: nodeArray{nodes: nodes}}
}

func (b *randomBalancer) Pick() Node {
	if len(b.nodes) == 0 {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	n := b.nodes[rand.Intn(len(b.nodes))]
	if n.Healthy() {
		return WrapNode{Node: n, nodeArrayRemove: b}
	}
	return nil
}
