package balancer

import (
	"sync"
)

type Balancer interface {
	Nodes() []Node
	Update(nodes []Node)
	Pick() Node
}

type Node interface {
	MarkFailure()  // mark node as failed
	Healthy() bool // health status, this method wont probe node health
	Weight() int   // weight
	Value() any
}

type nodeArray struct {
	mu    sync.RWMutex
	nodes []Node
}

func (na *nodeArray) Nodes() []Node {
	na.mu.RLock()
	defer na.mu.RUnlock()
	return na.nodes
}

func (na *nodeArray) updateLocked(nodes []Node) {
	if nodes == nil {
		nodes = []Node{}
	}
	na.nodes = nodes
}

func (na *nodeArray) Update(nodes []Node) {
	na.mu.Lock()
	defer na.mu.Unlock()
	na.updateLocked(nodes)
}

func (na *nodeArray) Remove(n Node) {
	na.mu.Lock()
	defer na.mu.Unlock()

	for i, node := range na.nodes {
		if node == n {
			copy(na.nodes[i:], na.nodes[i+1:])
			na.nodes[len(na.nodes)-1] = nil
			na.nodes = na.nodes[:len(na.nodes)-1]
			break
		}
	}
}

type nodeArrayRemove interface {
	Remove(Node)
}

type WrapNode struct {
	Node
	nodeArrayRemove
}

func (w WrapNode) MarkFailure() {
	w.nodeArrayRemove.Remove(w.Node)
	w.Node.MarkFailure()
}
