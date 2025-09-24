package syncx

import "sync"

type Map[T any] struct {
	mu sync.RWMutex
	m  map[string]T
}

func NewMap[T any]() *Map[T] {
	return &Map[T]{
		m: make(map[string]T),
	}
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *Map[T]) Range(f func(key string, value T) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.m {
		if !f(k, v) {
			break
		}
	}
}

// Keys returns a slice of all keys in the map.
// Keys are not sorted in any particular order.
// Keys are not cloned.
func (m *Map[T]) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice of all values in the map.
// Values are not sorted in any particular order.
// Values are not cloned.
func (m *Map[T]) Values() []T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]T, 0, len(m.m))
	for _, v := range m.m {
		values = append(values, v)
	}
	return values
}

// Clear removes all key/value pairs from the map.
func (m *Map[T]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m = map[string]T{}
}

// Len returns the number of elements in the map.
// Len is a constant-time operation.
func (m *Map[T]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.m)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map[T]) LoadOrStore(key string, value T) (actual T, loaded bool) {
	m.mu.RLock()
	v, ok := m.m[key]
	if ok {
		m.mu.RUnlock()
		return v, true
	}
	m.mu.RUnlock()

	m.mu.Lock()
	m.m[key] = value
	m.mu.Unlock()
	return value, false
}

// Load returns the value stored in the map for a key, or zero value if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *Map[T]) Load(key string) (T, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.m[key]
	return v, ok
}

// Store sets the value for a key.
func (m *Map[T]) Store(key string, value T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.m == nil {
		m.m = make(map[string]T)
	}
	m.m[key] = value
}

// Delete deletes the value for a key.
func (m *Map[T]) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, key)
}
