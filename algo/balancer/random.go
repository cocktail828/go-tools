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
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.Empty() {
		return nil
	}

	return b.nodes[rand.Intn(b.Len())]
}
