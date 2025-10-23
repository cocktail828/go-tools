package balancer

import (
	"sync"
)

type Balancer interface {
	// Nodes returns all nodes in balancer
	Nodes() []Node
	// Update updates balancer nodes
	Update(nodes []Node)
	// Pick picks a node from balancer
	// It returns nil if no node is available
	Pick() Node
}

type Node interface {
	// MarkFailure marks node as failed and remove it from balancer
	MarkFailure()
	// Healthy returns true if node is healthy, this method wont probe node health
	Healthy() bool
	// Weight returns node weight, only used by weight roundrobin
	Weight() int
	// Value returns node value
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

// Update updates balancer nodes
func (na *nodeArray) Update(nodes []Node) {
	tmp := map[any]struct{}{}
	for _, n := range nodes {
		tmp[n.Value()] = struct{}{}
	}

	na.mu.RLock()
	for _, n := range na.nodes {
		delete(tmp, n.Value())
	}
	na.mu.RUnlock()

	// nodes not changed
	if len(tmp) == 0 {
		return
	}

	na.mu.Lock()
	defer na.mu.Unlock()
	na.updateLocked(nodes)
}

func (na *nodeArray) Remove(node Node) {
	na.mu.Lock()
	defer na.mu.Unlock()

	for i, n := range na.nodes {
		if n == node {
			copy(na.nodes[i:], na.nodes[i+1:])
			na.nodes[len(na.nodes)-1] = nil
			na.nodes = na.nodes[:len(na.nodes)-1]
			break
		}
	}
}

type fallibleNode struct {
	Node
	*nodeArray
}

// MarkFailure marks node as unhealthy and remove it from balancer
func (w fallibleNode) MarkFailure() {
	w.nodeArray.Remove(w.Node)
	w.Node.MarkFailure()
}
