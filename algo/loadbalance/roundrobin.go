package loadbalance

import (
	"sync/atomic"
)

type Collection interface {
	// Validate will varify validity of element.
	Validate(int) bool
	// Len is the number of elements in the collection.
	Len() int
}

type RoundRobin struct {
	pos uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (lhs *RoundRobin) Get(c Collection) int {
	length := c.Len()
	for i := 0; i < length; i++ {
		pos := atomic.AddUint64(&lhs.pos, 1) % uint64(length)
		if c.Validate(int(pos)) {
			return int(pos)
		}
	}
	return -1
}
