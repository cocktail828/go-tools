package balancer

import "sync"

type rrBalancer struct {
	mu    sync.RWMutex
	pos   int
	array []Validator
}

func NewRR() Balancer {
	return &rrBalancer{}
}

func (b *rrBalancer) Update(array []Validator) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.array = array
}

func (b *rrBalancer) Pick() Validator {
	b.mu.RLock()
	defer b.mu.RUnlock()

	length := len(b.array)
	for i := 0; i < length; i++ {
		c := b.array[b.pos]
		if !c.IsOK() {
			continue
		}
		b.pos = (b.pos + 1) % length
		return c
	}
	return nil
}
