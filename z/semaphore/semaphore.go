package semaphore

import (
	"context"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
)

type Weighted struct {
	weighted *semaphore.Weighted
	size     int64        // the number of tokens
	assigned atomic.Int64 // the number of tokens assigned
}

func NewWeighted(n int64) *Weighted {
	return &Weighted{
		weighted: semaphore.NewWeighted(n),
		size:     n,
	}
}

func (s *Weighted) Acquire(ctx context.Context, n int64) error {
	if err := s.weighted.Acquire(ctx, n); err != nil {
		return err
	}
	s.assigned.Add(n)
	return nil
}

func (s *Weighted) Release(n int64) {
	s.weighted.Release(n)
	s.assigned.Add(-n)
}

func (s *Weighted) TryAcquire(n int64) bool {
	v := s.weighted.TryAcquire(n)
	if v {
		s.assigned.Add(n)
	}
	return v
}

func (s *Weighted) Assigned() int64 {
	return s.assigned.Load()
}

func (s *Weighted) Size() int64 {
	return s.size
}
