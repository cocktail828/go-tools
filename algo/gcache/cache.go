package gcache

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type EvictType string

const (
	TYPE_SIMPLE EvictType = "simple"
	TYPE_LRU    EvictType = "lru"
	TYPE_LFU    EvictType = "lfu"
	TYPE_ARC    EvictType = "arc"
)

var KeyNotFoundError = errors.New("Key not found.")

type Cache interface {
	EvictType() EvictType
	// Set inserts or updates the specified key-value pair.
	Set(key, value any) error
	// SetWithExpire inserts or updates the specified key-value pair with an expiration time.
	SetWithExpire(key, value any, expiration time.Duration) error
	// Get returns the value for the specified key if it is present in the cache.
	// If the key is not present in the cache and the cache has LoaderFunc,
	// invoke the `LoaderFunc` function and inserts the key-value pair in the cache.
	// If the key is not present in the cache and the cache does not have a LoaderFunc,
	// return KeyNotFoundError.
	Get(key any) (any, error)
	// GetAll returns a map containing all key-value pairs in the cache.
	GetALL(checkExpired bool) map[any]any
	get(key any, onLoad bool) (any, error)
	// Remove removes the specified key from the cache if the key is present.
	// Returns true if the key was present and the key has been deleted.
	Remove(key any) bool
	// Purge removes all key-value pairs from the cache.
	Purge()
	// Keys returns a slice containing all keys in the cache.
	Keys(checkExpired bool) []any
	// Len returns the number of items in the cache.
	Len(checkExpired bool) int
	// Has returns true if the key exists in the cache.
	Has(key any) bool

	statsAccessor
}

type baseCache struct {
	evictType        EvictType
	clock            Clock
	size             int
	loaderExpireFunc LoaderExpireFunc
	evictedFunc      EvictedFunc
	purgeVisitorFunc PurgeVisitorFunc
	addedFunc        AddedFunc
	deserializeFunc  DeserializeFunc
	serializeFunc    SerializeFunc
	expiration       *time.Duration
	mu               sync.RWMutex
	loadGroup        singleflight.Group
	cache            Cache
	*stats
}

type (
	LoaderFunc       func(any) (any, error)
	LoaderExpireFunc func(any) (any, *time.Duration, error)
	EvictedFunc      func(any, any)
	PurgeVisitorFunc func(any, any)
	AddedFunc        func(any, any)
	DeserializeFunc  func(any, any) (any, error)
	SerializeFunc    func(any, any) (any, error)
)

type CacheBuilder struct {
	clock            Clock
	evictType        EvictType
	size             int
	loaderExpireFunc LoaderExpireFunc
	evictedFunc      EvictedFunc
	purgeVisitorFunc PurgeVisitorFunc
	addedFunc        AddedFunc
	expiration       *time.Duration
	deserializeFunc  DeserializeFunc
	serializeFunc    SerializeFunc
}

func New(size int) *CacheBuilder {
	return &CacheBuilder{
		clock:     NewRealClock(),
		evictType: TYPE_SIMPLE,
		size:      size,
	}
}

func (cb *CacheBuilder) Clock(clock Clock) *CacheBuilder {
	cb.clock = clock
	return cb
}

// Set a loader function.
// loaderFunc: create a new value with this function if cached value is expired.
func (cb *CacheBuilder) LoaderFunc(loaderFunc LoaderFunc) *CacheBuilder {
	cb.loaderExpireFunc = func(k any) (any, *time.Duration, error) {
		v, err := loaderFunc(k)
		return v, nil, err
	}
	return cb
}

// Set a loader function with expiration.
// loaderExpireFunc: create a new value with this function if cached value is expired.
// If nil returned instead of time.Duration from loaderExpireFunc than value will never expire.
func (cb *CacheBuilder) LoaderExpireFunc(loaderExpireFunc LoaderExpireFunc) *CacheBuilder {
	cb.loaderExpireFunc = loaderExpireFunc
	return cb
}

func (cb *CacheBuilder) EvictType(tp EvictType) *CacheBuilder {
	cb.evictType = tp
	return cb
}

func (cb *CacheBuilder) Simple() *CacheBuilder {
	cb.evictType = TYPE_SIMPLE
	return cb
}

func (cb *CacheBuilder) LRU() *CacheBuilder {
	cb.evictType = TYPE_LRU
	return cb
}

func (cb *CacheBuilder) LFU() *CacheBuilder {
	cb.evictType = TYPE_LFU
	return cb
}

func (cb *CacheBuilder) ARC() *CacheBuilder {
	cb.evictType = TYPE_ARC
	return cb
}

func (cb *CacheBuilder) EvictedFunc(evictedFunc EvictedFunc) *CacheBuilder {
	cb.evictedFunc = evictedFunc
	return cb
}

func (cb *CacheBuilder) PurgeVisitorFunc(purgeVisitorFunc PurgeVisitorFunc) *CacheBuilder {
	cb.purgeVisitorFunc = purgeVisitorFunc
	return cb
}

func (cb *CacheBuilder) AddedFunc(addedFunc AddedFunc) *CacheBuilder {
	cb.addedFunc = addedFunc
	return cb
}

func (cb *CacheBuilder) DeserializeFunc(deserializeFunc DeserializeFunc) *CacheBuilder {
	cb.deserializeFunc = deserializeFunc
	return cb
}

func (cb *CacheBuilder) SerializeFunc(serializeFunc SerializeFunc) *CacheBuilder {
	cb.serializeFunc = serializeFunc
	return cb
}

func (cb *CacheBuilder) Expiration(expiration time.Duration) *CacheBuilder {
	cb.expiration = &expiration
	return cb
}

func (cb *CacheBuilder) Build() Cache {
	if cb.size <= 0 && cb.evictType != TYPE_SIMPLE {
		panic("gcache: Cache size <= 0")
	}

	return cb.build()
}

func (cb *CacheBuilder) build() Cache {
	switch cb.evictType {
	case TYPE_SIMPLE:
		return newSimpleCache(cb)
	case TYPE_LRU:
		return newLRUCache(cb)
	case TYPE_LFU:
		return newLFUCache(cb)
	case TYPE_ARC:
		return newARC(cb)
	default:
		panic("gcache: Unknown type " + cb.evictType)
	}
}

func buildCache(c *baseCache, cb *CacheBuilder) {
	c.evictType = cb.evictType
	c.clock = cb.clock
	c.size = cb.size
	c.loaderExpireFunc = cb.loaderExpireFunc
	c.expiration = cb.expiration
	c.addedFunc = cb.addedFunc
	c.deserializeFunc = cb.deserializeFunc
	c.serializeFunc = cb.serializeFunc
	c.evictedFunc = cb.evictedFunc
	c.purgeVisitorFunc = cb.purgeVisitorFunc
	c.stats = &stats{}
}

// load a new value using by specified key.
func (c *baseCache) load(key any, cb func(any, *time.Duration, error) (any, error)) (any, error) {
	v, err, _ := c.loadGroup.Do(fmt.Sprintf("%v", key), func() (v any, e error) {
		v, err := c.cache.get(key, true)
		if err == nil {
			return v, nil
		}

		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("Loader panics: %v", r)
			}
		}()
		return cb(c.loaderExpireFunc(key))
	})
	return v, err
}

func (c *baseCache) EvictType() EvictType {
	return c.evictType
}
