package balancer

import (
	"sync"
	"sync/atomic"

	"github.com/cocktail828/go-tools/z"
)

type failoverBalancer struct {
	mu    sync.RWMutex
	pos   atomic.Uint32
	array []Node
}

func NewFailover(array []Node) Balancer {
	return &failoverBalancer{array: array}
}

func (b *failoverBalancer) Pick() (n Node) {
	var array []Node
	z.WithRLock(&b.mu, func() { array = b.array })

	for i := b.pos.Load(); i < uint32(len(array)); i++ {
		n := array[i]
		if h, ok := n.(Healthy); ok && h.Healthy() {
			b.pos.Store(i)
			return n
		}
	}
	return nil
}
