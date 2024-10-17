package ttlmap

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrNoEntry = errors.New("no such entry")
)

type validateFunc func() bool

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

func (c *Cache[T]) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = map[string]entry[T]{}
}

func (c *Cache[T]) Set(key string, val T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = entry[T]{
		val:      val,
		validate: func() bool { return true },
	}
}

func (c *Cache[T]) SetWithTTL(key string, val T, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	c.cache[key] = entry[T]{
		val:      val,
		validate: func() bool { return time.Since(now) < ttl },
	}
}

func (c *Cache[T]) Get(key string) (T, error) {
	var a T
	c.mu.RLock()
	val, ok := c.cache[key]
	c.mu.RUnlock()
	if !ok || !val.validate() {
		return a, errors.WithMessagef(ErrNoEntry, "fn:Get, key:%q nonexist", key)
	}
	return val.val, nil
}

type Getter[T any] func(c *Cache[T], key string) (T, error)

func (c *Cache[T]) Fetch(key string, getter Getter[T]) (T, error) {
	var a T
	c.mu.RLock()
	val, ok := c.cache[key]
	c.mu.RUnlock()
	if !ok || !val.validate() {
		if getter == nil {
			return a, errors.WithMessagef(ErrNoEntry, "fn:Fetch, key:%q nonexist", key)
		}
		return getter(c, key)
	}
	return val.val, nil
}

func (c *Cache[T]) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

func (c *Cache[T]) Exist(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.cache[key]
	return ok && val.validate()
}
