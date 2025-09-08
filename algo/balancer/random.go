package balancer

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type randomBalancer struct {
	mu    sync.RWMutex
	array []Node
}

func NewRandom(array []Node) Balancer {
	return &randomBalancer{array: array}
}

func (b *randomBalancer) Pick() (n Node) {
	b.mu.RLock()
	array := b.array
	b.mu.RUnlock()

	if len(array) == 0 {
		return nil
	}
	c := array[rand.Intn(len(array))]
	return c
}
