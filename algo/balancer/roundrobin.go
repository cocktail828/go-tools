package balancer

import (
	"sync/atomic"
)

type Validator interface {
	// Validate will varify validity of element.
	Validate(int) bool
}

type Collection interface {
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
		if f, ok := c.(Validator); ok && !f.Validate(int(pos)) {
			continue
		}
		return int(pos)
	}
	return -1
}
