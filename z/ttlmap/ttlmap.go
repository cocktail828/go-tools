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
var expireFunc = func(ttl time.Duration) validateFunc {
	now := time.Now()
	return func() bool { return time.Since(now) < ttl }
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
		cache: map[string]entry[T]{},
	}
}

func (c *Cache[T]) Set(key string, val T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = entry[T]{
		val:      val,
		validate: alwaysTrue,
	}
}

func (c *Cache[T]) SetWithTTL(key string, val T, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = entry[T]{
		val:      val,
		validate: expireFunc(ttl),
	}
}

func (c *Cache[T]) Get(key string) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var a T
	val, ok := c.cache[key]
	if !ok || !val.validate() {
		return a, ErrNoEntry
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
