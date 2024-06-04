package set

import "sync"

type Set struct {
	mu      *sync.RWMutex
	members map[any]struct{}
}

func New() *Set {
	return &Set{
		mu:      &sync.RWMutex{},
		members: make(map[any]struct{}),
	}
}

func (s Set) Push(m any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[m] = struct{}{}
}

func (s Set) Pop(m any) any {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := s.members[m]
	delete(s.members, m)
	return v
}

func (s Set) Has(m any) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.members[m]
	return ok
}

func (s Set) Empty(m any) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.members) == 0
}

func (s *Set) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members = map[any]struct{}{}
}
