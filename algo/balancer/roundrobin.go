package balancer

import (
	"sync"

	"github.com/cocktail828/go-tools/z"
)

type rrBalancer struct {
	sync.RWMutex
	pos   uint16
	array []Node
}

func NewRR() Balancer {
	return &rrBalancer{}
}

func (b *rrBalancer) Update(array []Node) {
	z.WithLock(b, func() { b.array = array })
}

func (b *rrBalancer) Pick() (n Node) {
	var array []Node
	z.WithRLock(b, func() { array = b.array })

	if len(array) == 0 {
		return nil
	}
	c := array[b.pos%uint16(len(array))]
	b.pos++
	return c
}
