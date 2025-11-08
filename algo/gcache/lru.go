package gcache

import (
	"container/list"
	"time"
)

// Discards the least recently used items first.
type LRUCache struct {
	baseCache
	items     map[any]*list.Element
	evictList *list.List
}

func newLRUCache(cb *CacheBuilder) *LRUCache {
	c := &LRUCache{}
	buildCache(&c.baseCache, cb)
	c.cache = c
	c.init()
	return c
}

func (c *LRUCache) init() {
	c.evictList = list.New()
	c.items = make(map[any]*list.Element, c.size+1)
}

func (c *LRUCache) set(key, value any) (any, error) {
	var err error
	if c.serializeFunc != nil {
		value, err = c.serializeFunc(key, value)
		if err != nil {
			return nil, err
		}
	}

	// Check for existing item
	var item *lruItem
	if it, ok := c.items[key]; ok {
		c.evictList.MoveToFront(it)
		item = it.Value.(*lruItem)
		item.value = value
	} else {
		// Verify size not exceeded
		if c.evictList.Len() >= c.size {
			c.evict(1)
		}
		item = &lruItem{
			clock: c.clock,
			key:   key,
			value: value,
		}
		c.items[key] = c.evictList.PushFront(item)
	}

	if c.expiration != nil {
		t := c.clock.Now().Add(*c.expiration)
		item.expiration = &t
	}

	if c.addedFunc != nil {
		c.addedFunc(key, value)
	}

	return item, nil
}

// set a new key-value pair
func (c *LRUCache) Set(key, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.set(key, value)
	return err
}

// Set a new key-value pair with an expiration time
func (c *LRUCache) SetWithExpire(key, value any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, err := c.set(key, value)
	if err != nil {
		return err
	}

	t := c.clock.Now().Add(expiration)
	item.(*lruItem).expiration = &t
	return nil
}

// Get a value from cache pool using key if it exists.
// If it does not exists key and has LoaderFunc,
// generate a value using `LoaderFunc` method returns value.
func (c *LRUCache) Get(key any) (any, error) {
	v, err := c.get(key, false)
	if err == ErrKeyNotFoundError {
		return c.getWithLoader(key)
	}
	return v, err
}

func (c *LRUCache) get(key any, onLoad bool) (any, error) {
	v, err := c.getValue(key, onLoad)
	if err != nil {
		return nil, err
	}
	if c.deserializeFunc != nil {
		return c.deserializeFunc(key, v)
	}
	return v, nil
}

func (c *LRUCache) getValue(key any, onLoad bool) (any, error) {
	c.mu.Lock()
	item, ok := c.items[key]
	if ok {
		it := item.Value.(*lruItem)
		if !it.IsExpired(nil) {
			c.evictList.MoveToFront(item)
			v := it.value
			c.mu.Unlock()
			if !onLoad {
				c.stats.IncrHitCount()
			}
			return v, nil
		}
		c.removeElement(item)
	}
	c.mu.Unlock()
	if !onLoad {
		c.stats.IncrMissCount()
	}
	return nil, ErrKeyNotFoundError
}

func (c *LRUCache) getWithLoader(key any) (any, error) {
	if c.loaderExpireFunc == nil {
		return nil, ErrKeyNotFoundError
	}
	return c.load(key, func(v any, expiration *time.Duration, e error) (any, error) {
		if e != nil {
			return nil, e
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		item, err := c.set(key, v)
		if err != nil {
			return nil, err
		}
		if expiration != nil {
			t := c.clock.Now().Add(*expiration)
			item.(*lruItem).expiration = &t
		}
		return v, nil
	})
}

// evict removes the oldest item from the cache.
func (c *LRUCache) evict(count int) {
	for i := 0; i < count; i++ {
		ent := c.evictList.Back()
		if ent == nil {
			return
		} else {
			c.removeElement(ent)
		}
	}
}

// Has checks if key exists in cache
func (c *LRUCache) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	return c.has(key, &now)
}

func (c *LRUCache) has(key any, now *time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.Value.(*lruItem).IsExpired(now)
}

// Remove removes the provided key from the cache.
func (c *LRUCache) Remove(key any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *LRUCache) remove(key any) bool {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

func (c *LRUCache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	entry := e.Value.(*lruItem)
	delete(c.items, entry.key)
	if c.evictedFunc != nil {
		entry := e.Value.(*lruItem)
		c.evictedFunc(entry.key, entry.value)
	}
}

// GetALL returns all key-value pairs in the cache.
func (c *LRUCache) GetALL(checkExpired bool) map[any]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make(map[any]any, len(c.items))
	now := time.Now()
	for k, item := range c.items {
		if !checkExpired || c.has(k, &now) {
			items[k] = item.Value.(*lruItem).value
		}
	}
	return items
}

// Keys returns a slice of the keys in the cache.
func (c *LRUCache) Keys(checkExpired bool) []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]any, 0, len(c.items))
	now := time.Now()
	for k := range c.items {
		if !checkExpired || c.has(k, &now) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRUCache) Len(checkExpired bool) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !checkExpired {
		return len(c.items)
	}
	var length int
	now := time.Now()
	for k := range c.items {
		if c.has(k, &now) {
			length++
		}
	}
	return length
}

// Completely clear the cache
func (c *LRUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.purgeVisitorFunc != nil {
		for key, item := range c.items {
			it := item.Value.(*lruItem)
			v := it.value
			c.purgeVisitorFunc(key, v)
		}
	}

	c.init()
}

type lruItem struct {
	clock      Clock
	key        any
	value      any
	expiration *time.Time
}

// IsExpired returns boolean value whether this item is expired or not.
func (it *lruItem) IsExpired(now *time.Time) bool {
	if it.expiration == nil {
		return false
	}
	if now == nil {
		t := it.clock.Now()
		now = &t
	}
	return it.expiration.Before(*now)
}
