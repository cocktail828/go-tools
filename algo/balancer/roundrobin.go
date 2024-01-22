package balancer

import (
	"sync/atomic"
)

type roundRobinBuilder struct{}

func (roundRobinBuilder) Build() Balancer {
	return NewRoundRobin()
}

type roundRobin struct {
	pos   uint64
	array []any
}

var _ Balancer = &roundRobin{}

func init() {
	Register("round-robin", roundRobinBuilder{})
	Register("rr", roundRobinBuilder{})
}

func NewRoundRobin() *roundRobin {
	return &roundRobin{}
}

func (b *roundRobin) Update(array []any) error {
	b.array = array
	return nil
}

func (b *roundRobin) Pick() any {
	array := b.array
	length := len(array)
	for i := 0; i < length; i++ {
		pos := atomic.AddUint64(&b.pos, 1) % uint64(length)
		c := array[pos]
		if f, ok := c.(Validator); ok && !f.IsOK() {
			continue
		}
		return c
	}
	return nil
}
