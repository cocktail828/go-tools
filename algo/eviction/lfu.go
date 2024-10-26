package eviction

import (
	"container/heap"
	"sync"
	"time"
)

const (
	DefaultMaxFreq = 2 << 19
)

// LFUCache implements a Least Frequently Used cache with optional expiration.
type LFUCache struct {
	cache
	maxFreq  uint
	mu       sync.RWMutex
	items    map[string]*lfuItem
	freqHeap freqHeap // Min-heap based on frequency.
}

type lfuItem struct {
	key       string
	value     any
	freqEntry *freqEntry // Reference to the item's frequency entry
	expireAt  time.Time
}

type freqEntry struct {
	freq      uint
	items     map[*lfuItem]struct{}
	heapIndex int // Index in the heap for efficient updates
}

type freqHeap []*freqEntry

func (h freqHeap) Len() int           { return len(h) }
func (h freqHeap) Less(i, j int) bool { return h[i].freq < h[j].freq }
func (h freqHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].heapIndex = i
	h[j].heapIndex = j
}

func (h *freqHeap) Push(x any) {
	n := len(*h)
	entry := x.(*freqEntry)
	entry.heapIndex = n
	*h = append(*h, entry)
}

func (h *freqHeap) Pop() any {
	old := *h
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil // Avoid memory leak
	entry.heapIndex = -1
	*h = old[0 : n-1]
	return entry
}

func NewLFUCache(size uint, opts ...Option) Eviction {
	c := cache{
		size:    size,
		onEvict: func(s string, a any) {},
		onPurge: func(s string, a any) {},
	}
	for _, o := range opts {
		o(&c)
	}

	lfu := &LFUCache{
		cache:    c,
		maxFreq:  DefaultMaxFreq,
		items:    make(map[string]*lfuItem, size),
		freqHeap: freqHeap{},
	}
	heap.Init(&lfu.freqHeap)
	return lfu
}

func (c *LFUCache) SetMaxFreq(max uint) {
	if max > 0 {
		c.maxFreq = max
	}
}

func (c *LFUCache) Decay() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, item := range c.items {
		c.decayFrequency(item)
	}
}

func (c *LFUCache) Set(key string, value any) {
	c.SetWithExpiration(key, value, c.expiration)
}

func (c *LFUCache) SetWithExpiration(key string, value any, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	if item, exists := c.items[key]; exists {
		item.value = value
		item.expireAt = exp
		c.incrementFrequency(item)
	} else {
		if uint(len(c.items)) >= c.size {
			c.evict(1)
		}
		item := &lfuItem{key: key, value: value, expireAt: exp}
		c.items[key] = item
		c.addItemToFrequency(item, 0)
	}
}

func (c *LFUCache) addItemToFrequency(item *lfuItem, freq uint) {
	var entry *freqEntry
	// Try finding the frequency entry in the heap
	for _, e := range c.freqHeap {
		if e.freq == freq {
			entry = e
			break
		}
	}
	// Create a new frequency entry if none found
	if entry == nil {
		entry = &freqEntry{
			freq:  freq,
			items: make(map[*lfuItem]struct{}),
		}
		heap.Push(&c.freqHeap, entry)
	}
	// Add item to frequency entry
	entry.items[item] = struct{}{}
	item.freqEntry = entry
}

func (c *LFUCache) decayFrequency(item *lfuItem) {
	oldEntry := item.freqEntry
	newFreq := oldEntry.freq / 2
	c.addItemToFrequency(item, newFreq)

	// Remove item from old frequency entry and delete entry if empty
	delete(oldEntry.items, item)
	if len(oldEntry.items) == 0 {
		heap.Remove(&c.freqHeap, oldEntry.heapIndex)
	}
}

func (c *LFUCache) incrementFrequency(item *lfuItem) {
	oldEntry := item.freqEntry
	newFreq := oldEntry.freq + 1
	if newFreq >= uint(c.maxFreq) {
		newFreq = c.maxFreq
	}
	c.addItemToFrequency(item, newFreq)

	// Remove item from old frequency entry and delete entry if empty
	delete(oldEntry.items, item)
	if len(oldEntry.items) == 0 {
		heap.Remove(&c.freqHeap, oldEntry.heapIndex)
	}
}

func (c *LFUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[key]; exists {
		if item.expireAt.IsZero() || time.Now().Before(item.expireAt) {
			c.incrementFrequency(item)
			return item.value, true
		}
		c.removeItem(item)
	}

	return nil, false
}

func (c *LFUCache) evict(count int) {
	now := time.Now()
	for _, entry := range c.freqHeap {
		for item := range entry.items {
			if item.expireAt.Before(now) {
				c.removeItem(item)
				count--
				if count == 0 {
					return
				}
			}
		}
	}

	for count > 0 && len(c.freqHeap) > 0 {
		entry := c.freqHeap[0]
		for item := range entry.items {
			c.removeItem(item)
			count--
			if count == 0 {
				return
			}
		}
		if len(entry.items) == 0 {
			heap.Pop(&c.freqHeap)
		}
	}
}

func (c *LFUCache) removeItem(item *lfuItem) {
	if entry := item.freqEntry; entry != nil {
		delete(entry.items, item)
		if len(entry.items) == 0 {
			heap.Remove(&c.freqHeap, entry.heapIndex)
		}
	}
	delete(c.items, item.key)
	if c.onEvict != nil {
		c.onEvict(item.key, item.value)
	}
}

func (c *LFUCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, exists := c.items[key]
	return exists && (item.expireAt.IsZero() || time.Now().Before(item.expireAt))
}

func (c *LFUCache) Remove(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, exists := c.items[key]; exists {
		c.removeItem(item)
		return true
	}
	return false
}

func (c *LFUCache) lookupWithCallbackLocked(includeExpired bool, cb func(*lfuItem)) {
	now := time.Now()
	for _, elem := range c.items {
		if includeExpired || elem.expireAt.IsZero() || now.Before(elem.expireAt) {
			cb(elem)
		}
	}
}

func (c *LFUCache) GetAll(includeExpired bool) map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]any, len(c.items))
	c.lookupWithCallbackLocked(includeExpired, func(li *lfuItem) {
		result[li.key] = li.value
	})
	return result
}

func (c *LFUCache) Keys(includeExpired bool) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.items))
	c.lookupWithCallbackLocked(includeExpired, func(li *lfuItem) {
		keys = append(keys, li.key)
	})
	return keys
}

func (c *LFUCache) Len(includeExpired bool) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := 0
	c.lookupWithCallbackLocked(includeExpired, func(li *lfuItem) {
		count++
	})
	return count
}

func (c *LFUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, item := range c.items {
		if c.onPurge != nil {
			c.onPurge(key, item.value)
		}
	}
	c.items = make(map[string]*lfuItem, c.size)
	c.freqHeap = freqHeap{}
	heap.Init(&c.freqHeap)
}
