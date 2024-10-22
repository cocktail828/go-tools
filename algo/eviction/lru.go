package eviction

import (
	"container/list"
	"sync"
	"time"
)

// Discards the least recently used items first.
type LRUCache struct {
	cache

	mu        sync.RWMutex
	items     map[string]*list.Element
	evictList *list.List
}

func NewLRUCache(size int) Eviction {
	c := &LRUCache{cache: cache{size: size}}
	c.init()
	return c
}

func (c *LRUCache) init() {
	c.evictList = list.New()
	c.items = make(map[string]*list.Element, c.size+1)
}

func (c *LRUCache) set(key string, value any, expiration time.Duration) {
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
			key:   key,
			value: value,
		}
		c.items[key] = c.evictList.PushFront(item)
	}

	item.addAt = time.Now()
	item.isExpired = func(now time.Time) bool {
		return now.Sub(item.addAt) >= expiration
	}
}

// set a new key-value pair
func (c *LRUCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value, c.expiration)
}

// Set a new key-value pair with an expiration time
func (c *LRUCache) SetWithExpire(key string, value any, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value, expiration)
}

// Get a value from cache pool using key if it exists.
func (c *LRUCache) Get(key string) (any, error) {
	return c.get(key)
}

func (c *LRUCache) get(key string) (any, error) {
	c.mu.Lock()
	item, ok := c.items[key]
	if ok {
		it := item.Value.(*lruItem)
		if !it.isExpired(time.Now()) {
			c.evictList.MoveToFront(item)
			v := it.value
			c.mu.Unlock()
			c.IncrHitCount()
			return v, nil
		}
		c.removeElement(item)
	}
	c.mu.Unlock()
	c.IncrMissCount()
	return nil, ErrKeyNotFound
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
func (c *LRUCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	return c.has(key, now)
}

func (c *LRUCache) has(key string, now time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.Value.(*lruItem).isExpired(now)
}

// Remove removes the provided key from the cache.
func (c *LRUCache) Remove(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *LRUCache) remove(key string) bool {
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
	if c.onEvict != nil {
		entry := e.Value.(*lruItem)
		c.onEvict(entry.key, entry.value)
	}
}

// GetALL returns all key-value pairs in the cache.
func (c *LRUCache) GetAll(checkExpired bool) map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make(map[string]any, len(c.items))
	now := time.Now()
	for k, item := range c.items {
		if !checkExpired || c.has(k, now) {
			items[k] = item.Value.(*lruItem).value
		}
	}
	return items
}

// Keys returns a slice of the keys in the cache.
func (c *LRUCache) Keys(checkExpired bool) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.items))
	now := time.Now()
	for k := range c.items {
		if !checkExpired || c.has(k, now) {
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
		if c.has(k, now) {
			length++
		}
	}
	return length
}

// Completely clear the cache
func (c *LRUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onPurge != nil {
		for key, item := range c.items {
			it := item.Value.(*lruItem)
			v := it.value
			c.onPurge(key, v)
		}
	}

	c.init()
}

type lruItem struct {
	key       string
	value     any
	addAt     time.Time
	isExpired func(time.Time) bool
}
