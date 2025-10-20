package balancer

import "sync"

// for failover balancer
type Healthy interface {
	Healthy() bool // health status
}

type Balancer interface {
	Update(nodes []Node)
	Pick() Node
}

type Node interface {
	Weight() int // weight
	Value() any
}

type nodeArray struct {
	mu    sync.RWMutex
	nodes []Node
}

func (ns *nodeArray) Nodes() []Node {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.nodes
}

func (ns *nodeArray) Len() int {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return len(ns.nodes)
}

func (ns *nodeArray) Empty() bool {
	return ns.Len() == 0
}

func (ns *nodeArray) updateLocked(nodes []Node) {
	if nodes == nil {
		nodes = []Node{}
	}
	ns.nodes = nodes
}

func (ns *nodeArray) Update(nodes []Node) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.updateLocked(nodes)
}
