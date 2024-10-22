package eviction

import (
	"container/list"
	"sync"
	"time"
)

// Discards the least frequently used items first.
type LFUCache struct {
	cache
	mu       sync.RWMutex
	items    map[string]*lfuItem
	freqList *list.List // list for freqEntry
}

type lfuItem struct {
	key         string
	value       any
	freqElement *list.Element
	addAt       time.Time
	isExpired   func(time.Time) bool
}

type freqEntry struct {
	freq  uint
	items map[*lfuItem]struct{}
}

func NewLFUCache(size int) Eviction {
	c := &LFUCache{cache: cache{size: size}}
	c.init()
	return c
}

func (c *LFUCache) init() {
	c.freqList = list.New()
	c.items = make(map[string]*lfuItem, c.size)
	c.freqList.PushFront(&freqEntry{
		freq:  0,
		items: make(map[*lfuItem]struct{}),
	})
}

// Set a new key-value pair
func (c *LFUCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value, c.expiration)
}

// Set a new key-value pair with an expiration time
func (c *LFUCache) SetWithExpire(key string, value any, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value, expiration)
}

func (c *LFUCache) set(key string, value any, expiration time.Duration) {
	// Check for existing item
	item, ok := c.items[key]
	if ok {
		item.value = value
	} else {
		// Verify size not exceeded
		if len(c.items) >= c.size {
			c.evict(1)
		}
		item = &lfuItem{
			key:         key,
			value:       value,
			freqElement: nil,
		}
		el := c.freqList.Front()
		fe := el.Value.(*freqEntry)
		fe.items[item] = struct{}{}

		item.freqElement = el
		c.items[key] = item
	}

	item.addAt = time.Now()
	item.isExpired = func(now time.Time) bool {
		return now.Sub(item.addAt) >= expiration
	}
}

// Get a value from cache pool using key if it exists.
func (c *LFUCache) Get(key string) (any, error) {
	return c.get(key)
}

func (c *LFUCache) get(key string) (any, error) {
	c.mu.Lock()
	item, ok := c.items[key]
	if ok {
		if !item.isExpired(time.Now()) {
			c.increment(item)
			v := item.value
			c.mu.Unlock()
			c.IncrHitCount()
			return v, nil
		}
		c.removeItem(item)
	}
	c.mu.Unlock()
	c.IncrMissCount()
	return nil, ErrKeyNotFound
}

func (c *LFUCache) increment(item *lfuItem) {
	currentFreqElement := item.freqElement
	currentFreqEntry := currentFreqElement.Value.(*freqEntry)
	nextFreq := currentFreqEntry.freq + 1
	delete(currentFreqEntry.items, item)

	// a boolean whether reuse the empty current entry
	removable := isRemovableFreqEntry(currentFreqEntry)

	// insert item into a valid entry
	nextFreqElement := currentFreqElement.Next()
	switch {
	case nextFreqElement == nil || nextFreqElement.Value.(*freqEntry).freq > nextFreq:
		if removable {
			currentFreqEntry.freq = nextFreq
			nextFreqElement = currentFreqElement
		} else {
			nextFreqElement = c.freqList.InsertAfter(&freqEntry{
				freq:  nextFreq,
				items: make(map[*lfuItem]struct{}),
			}, currentFreqElement)
		}
	case nextFreqElement.Value.(*freqEntry).freq == nextFreq:
		if removable {
			c.freqList.Remove(currentFreqElement)
		}
	default:
		panic("unreachable")
	}
	nextFreqElement.Value.(*freqEntry).items[item] = struct{}{}
	item.freqElement = nextFreqElement
}

// evict removes the least frequence item from the cache.
func (c *LFUCache) evict(count int) {
	entry := c.freqList.Front()
	for i := 0; i < count; {
		if entry == nil {
			return
		} else {
			for item := range entry.Value.(*freqEntry).items {
				if i >= count {
					return
				}
				c.removeItem(item)
				i++
			}
			entry = entry.Next()
		}
	}
}

// Has checks if key exists in cache
func (c *LFUCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	return c.has(key, now)
}

func (c *LFUCache) has(key string, now time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.isExpired(now)
}

// Remove removes the provided key from the cache.
func (c *LFUCache) Remove(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *LFUCache) remove(key string) bool {
	if item, ok := c.items[key]; ok {
		c.removeItem(item)
		return true
	}
	return false
}

// removeElement is used to remove a given list element from the cache
func (c *LFUCache) removeItem(item *lfuItem) {
	entry := item.freqElement.Value.(*freqEntry)
	delete(c.items, item.key)
	delete(entry.items, item)
	if isRemovableFreqEntry(entry) {
		c.freqList.Remove(item.freqElement)
	}
	if c.onEvict != nil {
		c.onEvict(item.key, item.value)
	}
}

// GetALL returns all key-value pairs in the cache.
func (c *LFUCache) GetAll(checkExpired bool) map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make(map[string]any, len(c.items))
	now := time.Now()
	for k, item := range c.items {
		if !checkExpired || c.has(k, now) {
			items[k] = item.value
		}
	}
	return items
}

// Keys returns a slice of the keys in the cache.
func (c *LFUCache) Keys(checkExpired bool) []string {
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
func (c *LFUCache) Len(checkExpired bool) int {
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
func (c *LFUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onPurge != nil {
		for key, item := range c.items {
			c.onPurge(key, item.value)
		}
	}

	c.init()
}

func isRemovableFreqEntry(entry *freqEntry) bool {
	return entry.freq != 0 && len(entry.items) == 0
}
