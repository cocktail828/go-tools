package cache

import (
	"context"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/pkg/kvstore"
	"github.com/cocktail828/go-tools/z/locker"
	"github.com/pkg/errors"

	"github.com/cespare/xxhash/v2"
)

var (
	ErrNoEntry = errors.New("no such entry")
	ErrInvalid = errors.New("invalid entry")
)

type ValidateFunc func() bool

var (
	AlwaysTrue = func() bool { return true }
	ExpireFunc = func(v time.Duration) ValidateFunc {
		if v <= 0 {
			return AlwaysTrue
		}
		now := time.Now()
		return func() bool {
			return time.Since(now) < v
		}
	}
)

type entry struct {
	Value    []byte
	Validate ValidateFunc
}

type Cache struct {
	mu      sync.RWMutex
	entries [1024]map[string]entry
}

func New() kvstore.KV {
	c := &Cache{}
	for idx := 0; idx < len(c.entries); idx++ {
		c.entries[idx] = make(map[string]entry)
	}
	return c
}

func (c *Cache) String() string { return "cache" }

func (c *Cache) Close() error { return nil }
func (c *Cache) Watch(...kvstore.Option) kvstore.Watcher {
	return kvstore.NopWatcher{}
}

func (c *Cache) index(key string) int { return int(xxhash.Sum64String(key) % uint64(len(c.entries))) }

func (c *Cache) setLocked(ctx context.Context, key string, val []byte) {
	en := entry{
		Value:    val,
		Validate: AlwaysTrue,
	}

	if v := ctx.Value(validateOption{}); v != nil {
		en.Validate = v.(ValidateFunc)
	}
	c.entries[c.index(key)][key] = en
}

func (c *Cache) getLocked(key string) ([]byte, error) {
	w, ok := c.entries[c.index(key)][key]
	if !ok {
		return nil, ErrNoEntry
	}
	if !w.Validate() {
		return nil, ErrInvalid
	}
	return w.Value, nil
}

func (c *Cache) Set(key string, val []byte, opts ...kvstore.Option) error {
	ctx := context.Background()
	for _, o := range opts {
		ctx = o(ctx)
	}

	locker.WithLock(&c.mu, func() { c.setLocked(ctx, key, val) })
	return nil
}

func (c *Cache) Get(key string, opts ...kvstore.Option) ([]kvstore.KVPair, error) {
	ctx := context.Background()
	for _, o := range opts {
		ctx = o(ctx)
	}

	var val []byte
	var err error
	locker.WithRLock(&c.mu, func() { val, err = c.getLocked(key) })
	if err == ErrNoEntry {
		if v := ctx.Value(readthroughOption{}); v != nil {
			val, err = v.(Getter)(key)
			if err != nil {
				return nil, errors.WithMessage(ErrNoEntry, err.Error())
			}

			locker.WithLock(&c.mu, func() { c.setLocked(ctx, key, val) })
		}
	}
	return []kvstore.KVPair{{key, val}}, err
}

func (c *Cache) Del(key string, opts ...kvstore.Option) error {
	locker.WithLock(&c.mu, func() { delete(c.entries[c.index(key)], key) })
	return nil
}

func (c *Cache) Evict() {
	locker.WithLock(&c.mu, func() {
		for idx := 0; idx < len(c.entries); idx++ {
			for k, w := range c.entries[idx] {
				if !w.Validate() {
					delete(c.entries[idx], k)
				}
			}
		}
	})
}
