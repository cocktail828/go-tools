package balancer

import (
	"math/rand"
	"time"
)

var (
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type randomBalancer struct {
	*nodeArray
}

func NewRandom(nodes []Node) Balancer {
	return &randomBalancer{nodeArray: &nodeArray{nodes: nodes}}
}

func (b *randomBalancer) String() string {
	return "random"
}

func (b *randomBalancer) Pick() Node {
	if len(b.nodes) == 0 {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for i := 0; i < len(b.nodes)/2; i++ {
		n := b.nodes[randGen.Intn(len(b.nodes))]
		if n.Healthy() {
			return fallibleNode{Node: n, nodeArray: b.nodeArray}
		}
	}
	return nil
}
