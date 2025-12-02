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

type Candidate struct {
	mu    sync.RWMutex
	nodes []Node
}

func (c *Candidate) Nodes() []Node {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodes
}

func (c *Candidate) updateLocked(nodes []Node) {
	if nodes == nil {
		nodes = []Node{}
	}
	c.nodes = nodes
}

// Update updates balancer nodes
func (c *Candidate) Update(nodes []Node) {
	tmp := map[any]struct{}{}
	for _, n := range nodes {
		tmp[n.Value()] = struct{}{}
	}

	c.mu.RLock()
	for _, n := range c.nodes {
		delete(tmp, n.Value())
	}
	c.mu.RUnlock()

	// nodes not changed
	if len(tmp) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.updateLocked(nodes)
}

func (c *Candidate) Remove(node Node) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, n := range c.nodes {
		if n == node {
			copy(c.nodes[i:], c.nodes[i+1:])
			c.nodes[len(c.nodes)-1] = nil
			c.nodes = c.nodes[:len(c.nodes)-1]
			break
		}
	}
}

type fallibleNode struct {
	Node
	*Candidate
}

// MarkFailure marks node as unhealthy and remove it from balancer
func (w fallibleNode) MarkFailure() {
	w.Candidate.Remove(w.Node)
	w.Node.MarkFailure()
}
