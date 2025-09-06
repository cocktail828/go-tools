package balancer

import (
	"sync"

	"github.com/cocktail828/go-tools/z"
)

type rrBalancer struct {
	mu    sync.RWMutex
	pos   uint16
	array []Node
}

func NewRR(array []Node) Balancer {
	return &rrBalancer{array: array}
}

func (b *rrBalancer) Pick() (n Node) {
	var array []Node
	z.WithLock(b.mu.RLocker(), func() { array = b.array })

	if len(array) == 0 {
		return nil
	}
	c := array[b.pos%uint16(len(array))]
	b.pos++
	return c
}
