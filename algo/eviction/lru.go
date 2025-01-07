package eviction

import (
	"container/list"
	"sync"
	"time"
)

type LRU struct {
	cache
	mu        sync.RWMutex
	items     map[string]*list.Element
	evictList *list.List
}

type lruItem struct {
	key      string
	value    any
	freq     uint
	expireAt time.Time
}

func NewLRU(size uint, opts ...Option) Eviction {
	c := cache{
		size:    size,
		onEvict: func(s string, a any) {},
		onPurge: func(s string, a any) {},
	}
	for _, o := range opts {
		o(&c)
	}

	return &LRU{
		cache:     c,
		items:     make(map[string]*list.Element, size),
		evictList: list.New(),
	}
}

func (c *LRU) Set(key string, value any) {
	c.SetWithExpiration(key, value, c.expiration)
}

func (c *LRU) SetWithExpiration(key string, value any, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	if elem, exists := c.items[key]; exists {
		c.evictList.MoveToFront(elem)
		item := elem.Value.(*lruItem)
		item.value = value
		item.expireAt = exp
	} else {
		if uint(c.evictList.Len()) >= c.size {
			c.evict(1)
		}
		c.items[key] = c.evictList.PushFront(&lruItem{
			key:      key,
			value:    value,
			expireAt: exp,
		})
	}
}

func (c *LRU) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		item := elem.Value.(*lruItem)
		now := time.Now()
		if item.expireAt.IsZero() || now.Before(item.expireAt) {
			c.evictList.MoveToFront(elem)
			item.freq++
			return item.value, true
		}
		c.removeElement(elem)
	}
	return nil, false
}

func (c *LRU) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, exists := c.items[key]
	return exists && (elem.Value.(*lruItem).expireAt.IsZero() || time.Now().Before(elem.Value.(*lruItem).expireAt))
}

func (c *LRU) Remove(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if exists {
		c.removeElement(elem)
		return true
	}
	return false
}

func (c *LRU) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onPurge != nil {
		for key, elem := range c.items {
			c.onPurge(key, elem.Value.(*lruItem).value)
		}
	}
	c.items = make(map[string]*list.Element, c.size)
	c.evictList.Init()
}

func (c *LRU) evict(count int) {
	now := time.Now()
	for _, elem := range c.items {
		item := elem.Value.(*lruItem)
		if !item.expireAt.IsZero() && !now.Before(item.expireAt) {
			count--
			if count == 0 {
				return
			}
		}
	}

	for i := 0; i < count; i++ {
		elem := c.evictList.Back()
		if elem != nil {
			c.removeElement(elem)
		}
	}
}

func (c *LRU) removeElement(e *list.Element) {
	item := e.Value.(*lruItem)
	if c.onEvict != nil {
		c.onEvict(item.key, item.value)
	}
	delete(c.items, item.key)
	c.evictList.Remove(e)
}

func (c *LRU) lookupWithCallbackLocked(includeExpired bool, cb func(*lruItem)) {
	now := time.Now()
	for _, elem := range c.items {
		item := elem.Value.(*lruItem)
		if includeExpired || (item.expireAt.IsZero() || now.Before(item.expireAt)) {
			cb(item)
		}
	}
}

func (c *LRU) GetAll(includeExpired bool) map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]any, len(c.items))
	c.lookupWithCallbackLocked(includeExpired, func(li *lruItem) {
		result[li.key] = li.value
	})
	return result
}

func (c *LRU) Keys(includeExpired bool) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0, len(c.items))
	c.lookupWithCallbackLocked(includeExpired, func(li *lruItem) {
		result = append(result, li.key)
	})
	return result
}

func (c *LRU) Len(includeExpired bool) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := 0
	c.lookupWithCallbackLocked(includeExpired, func(li *lruItem) {
		result++
	})
	return result
}

func (c *LRU) Frequency(key string) uint {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, exists := c.items[key]
	if exists {
		item := elem.Value.(*lruItem)
		if item.expireAt.IsZero() || time.Now().Before(item.expireAt) {
			return item.freq
		}
	}
	return 0
}
