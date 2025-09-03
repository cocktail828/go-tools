package hystrix

import (
	"sync"
)

type Assigner struct {
	mu        sync.RWMutex
	maxCount  int
	allocated int
}

func (s *Assigner) TryAcquire() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.allocated < s.maxCount {
		s.allocated++
		return true
	}
	return false
}

func (s *Assigner) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.allocated <= 0 {
		panic("cannot release unacquired assigner")
	}
	s.allocated--
}

func (s *Assigner) Resize(newMax int) {
	if newMax < 0 {
		panic("newMax must be non-negative")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxCount = newMax
}

func (s *Assigner) Allocated() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.allocated
}

func (s *Assigner) Available() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.maxCount - s.allocated
}
