package syncx

import "sync"

type Set[T comparable] struct {
	mu sync.RWMutex
	m  map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		m: make(map[T]struct{}),
	}
}

// Add adds a value to the set.
func (s *Set[T]) Add(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[v] = struct{}{}
}

// Remove removes a value from the set.
func (s *Set[T]) Remove(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, v)
}

// Has reports whether the set contains the value.
func (s *Set[T]) Has(v T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[v]
	return ok
}

// Clear removes all values from the set.
func (s *Set[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m = make(map[T]struct{})
}

// Len returns the number of elements in the set.
func (s *Set[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}

// Values returns a slice of all values in the set.
// Values are not sorted in any particular order.
// Values are not cloned.
func (s *Set[T]) Values() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]T, 0, len(s.m))
	for v := range s.m {
		values = append(values, v)
	}
	return values
}

// Range calls f sequentially for each value present in the set.
// If f returns false, range stops the iteration.
func (s *Set[T]) Range(f func(value T) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for v := range s.m {
		if !f(v) {
			break
		}
	}
}

// Set operations
// Union returns a new set that is the union of s and other.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	union := NewSet[T]()
	for v := range s.m {
		union.Add(v)
	}
	for v := range other.m {
		union.Add(v)
	}
	return union
}

// Intersection returns a new set that is the intersection of s and other.
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	intersection := NewSet[T]()
	for v := range s.m {
		if other.Has(v) {
			intersection.Add(v)
		}
	}
	return intersection
}

// Difference returns a new set that is the difference of s and other.
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	difference := NewSet[T]()
	for v := range s.m {
		if !other.Has(v) {
			difference.Add(v)
		}
	}
	return difference
}

// IsSubset returns true if s is a subset of other.
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	for v := range s.m {
		if !other.Has(v) {
			return false
		}
	}
	return true
}

// IsSuperset returns true if s is a superset of other.
func (s *Set[T]) IsSuperset(other *Set[T]) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	return other.IsSubset(s)
}
