package cache

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNoEntry = errors.New("no such entry")
	ErrInvalid = errors.New("invalid entry")
)

type ValidateFunc func() bool

var (
	AlwaysTrue = func() bool { return true }
	ExpireFunc = func(v time.Duration) ValidateFunc {
		now := time.Now()
		return func() bool {
			return time.Since(now) < v
		}
	}
)

type entry[T any] struct {
	Value    T
	Validate ValidateFunc
}

type Config[T any] struct {
	OnEviction func(string, T)
}

type Cache[T any] struct {
	Config[T]
	mu      sync.RWMutex
	entries map[string]*entry[T]
}

func New[T any](cfg Config[T]) *Cache[T] {
	if cfg.OnEviction == nil {
		cfg.OnEviction = func(s string, t T) {}
	}

	return &Cache[T]{
		Config:  cfg,
		entries: map[string]*entry[T]{},
	}
}

func (c *Cache[T]) setLocked(key string, val T, oo *OpOption) {
	en := entry[T]{
		Value:    val,
		Validate: AlwaysTrue,
	}

	if v := oo.context.Value(validateOption{}); v != nil {
		en.Validate = v.(ValidateFunc)
	}

	if old, ok := c.entries[key]; ok {
		c.OnEviction(key, old.Value)
	}
	c.entries[key] = &en
}

func (c *Cache[T]) getLocked(key string, _ *OpOption) (T, error) {
	var a T
	w, ok := c.entries[key]
	if !ok {
		return a, ErrNoEntry
	}
	if !w.Validate() {
		return a, ErrInvalid
	}
	return w.Value, nil
}

func (c *Cache[T]) Set(key string, val T, opts ...Option) {
	oo := NewOpOption(opts...)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setLocked(key, val, oo)
}

func (c *Cache[T]) Get(key string, opts ...Option) (val T, err error) {
	oo := NewOpOption(opts...)
	func() {
		c.mu.RLock()
		defer c.mu.RUnlock()
		val, err = c.getLocked(key, oo)
	}()

	if v := oo.context.Value(evictOption{}); v != nil {
		c.Del(key)
	}
	return
}

func (c *Cache[T]) GetSet(key string, val T, opts ...Option) (T, error) {
	oo := NewOpOption(opts...)
	c.mu.Lock()
	defer c.mu.Unlock()
	old, err := c.getLocked(key, oo)
	c.setLocked(key, val, oo)
	return old, err
}

func (c *Cache[T]) Del(key string, opts ...Option) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if old, ok := c.entries[key]; ok {
		c.OnEviction(key, old.Value)
	}
	delete(c.entries, key)
}

func (c *Cache[T]) Expire(key string, ttl time.Duration, opts ...Option) {
	oo := NewOpOption(opts...)
	WithValidate(ExpireFunc(ttl))(oo)

	c.mu.Lock()
	defer c.mu.Unlock()
	if val, ok := c.entries[key]; ok && val.Validate() {
		val.Validate = oo.context.Value(validateOption{}).(ValidateFunc)
	}
}

func (c *Cache[T]) Evict() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, w := range c.entries {
		if !w.Validate() {
			c.OnEviction(k, w.Value)
			delete(c.entries, k)
		}
	}
}
