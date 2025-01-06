package ttlmap

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNoEntry = errors.New("no such entry")
)

type validateFunc func() bool

var alwaysTrue = func() bool { return true }

// expireFunc returns a function that checks if the current time is before the expiration time.
func expireFunc(expiration time.Time) validateFunc {
	return func() bool { return time.Now().Before(expiration) }
}

type entry[T any] struct {
	val      T
	validate validateFunc
}

type Cache[T any] struct {
	mu    sync.RWMutex
	cache map[string]entry[T]
}

func New[T any]() *Cache[T] {
	return &Cache[T]{
		cache: make(map[string]entry[T]),
	}
}

// setEntry is a helper function to set an entry in the cache.
func (c *Cache[T]) setEntry(key string, val T, validate validateFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = entry[T]{
		val:      val,
		validate: validate,
	}
}

func (c *Cache[T]) Set(key string, val T) {
	c.setEntry(key, val, alwaysTrue)
}

func (c *Cache[T]) SetWithTTL(key string, val T, ttl time.Duration) {
	c.setEntry(key, val, expireFunc(time.Now().Add(ttl)))
}

func (c *Cache[T]) Get(key string) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.cache[key]
	if !ok || !val.validate() {
		var zero T
		return zero, ErrNoEntry
	}
	return val.val, nil
}

// GetDel retrieves the value for a key and deletes it if it's invalid.
func (c *Cache[T]) GetDel(key string) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer delete(c.cache, key)

	val, ok := c.cache[key]
	if !ok || !val.validate() {
		var zero T
		return zero, ErrNoEntry
	}
	return val.val, nil
}

func (c *Cache[T]) Del(key string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	old, ok := c.cache[key]
	delete(c.cache, key)
	return old.val, ok
}
