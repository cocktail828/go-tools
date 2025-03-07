package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	// Epoch is set to the Twitter snowflake epoch of Nov 04 2010 01:42:54 UTC in milliseconds.
	// You may customize this to set a different epoch for your application.
	Epoch int64 = 1288834974657

	// NodeBits holds the number of bits to use for Node.
	// Remember, you have a total of 22 bits to share between Node/Step.
	NodeBits uint8 = 10

	// StepBits holds the number of bits to use for Step.
	// Remember, you have a total of 22 bits to share between Node/Step.
	StepBits uint8 = 12

	// Pre-calculate constants
	maxNode   = -1 ^ (-1 << NodeBits)
	maxStep   = -1 ^ (-1 << StepBits)
	timeShift = NodeBits + StepBits
	nodeShift = StepBits
)

// A Node struct holds the basic information needed for a snowflake generator node.
type Node struct {
	mu    sync.Mutex
	epoch time.Time
	time  int64
	node  int64
	step  int64
}

// NewNode returns a new snowflake node that can be used to generate snowflake IDs.
func NewNode(node int64) (*Node, error) {
	if node < 0 || node > maxNode {
		return nil, fmt.Errorf("node number must be between 0 and %d", maxNode)
	}

	return &Node{
		epoch: time.UnixMilli(Epoch),
		node:  node,
	}, nil
}

// Generate creates and returns a unique snowflake ID.
// To help guarantee uniqueness:
// - Make sure your system is keeping accurate system time.
// - Make sure you never have multiple nodes running with the same node ID.
func (n *Node) Generate() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Since(n.epoch).Milliseconds()
	if now == n.time {
		n.step = (n.step + 1) & maxStep
		if n.step == 0 {
			for now <= n.time {
				now = time.Since(n.epoch).Milliseconds()
			}
		}
	} else {
		n.step = 0
	}

	n.time = now
	return (now << timeShift) | (n.node << nodeShift) | (n.step)
}
