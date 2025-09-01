package balancer

import (
	"math/rand"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/z"
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
	var array []Node
	z.WithRLock(&b.mu, func() { array = b.array })

	if len(array) == 0 {
		return nil
	}
	c := array[rand.Intn(len(array))]
	return c
}
