package eviction

import (
	"time"
)

// WindowLFU combines LFU and LRU caches to form a hybrid cache.
type WindowLFU struct {
	lru                *LRU
	lfu                *LFU
	windowSize         uint // Size of the LRU window
	totalSize          uint // Total size of the cache
	promotionThreshold uint // Number of accesses before promotion to LFU
}

// NewWindowLFUCache initializes a new Window-LFU cache with LRU and LFU regions.
func NewWindowLFU(windowSize, totalSize, promotionThreshold uint, opts ...Option) Eviction {
	if windowSize >= totalSize {
		panic("windowSize should be less than totalSize")
	}

	lru := NewLRU(windowSize, opts...).(*LRU)
	lfu := NewLFU(totalSize-windowSize, opts...).(*LFU)

	return &WindowLFU{
		lru:                lru,
		lfu:                lfu,
		windowSize:         windowSize,
		totalSize:          totalSize,
		promotionThreshold: promotionThreshold,
	}
}

func (w *WindowLFU) Set(key string, value any) {
	w.SetWithTTL(key, value, w.lfu.expiration)
}

// Set adds a key-value pair to the cache with optional expiration.
func (w *WindowLFU) SetWithTTL(key string, value any, expiration time.Duration) {
	// Check if already in LFU
	if _, found := w.lfu.Get(key); found {
		w.lfu.SetWithTTL(key, value, expiration)
		return
	}

	// Add to LRU if not in LFU
	w.lru.SetWithTTL(key, value, expiration)
}

// Get retrieves a value by key, promoting it to LFU if it meets the access threshold.
func (w *WindowLFU) Get(key string) (any, bool) {
	// Check in LFU first
	if value, found := w.lfu.Get(key); found {
		return value, true
	}

	// Check in LRU and promote if access threshold is reached
	if value, found := w.lru.Get(key); found {
		// If frequency threshold met, promote to LFU
		if w.lru.Frequency(key) >= w.promotionThreshold {
			w.promoteToLFU(key, value)
		}
		return value, true
	}

	return nil, false
}

// promoteToLFU moves an item from LRU to LFU cache.
func (w *WindowLFU) promoteToLFU(key string, value any) {
	// Remove from LRU
	w.lru.Remove(key)

	// Add to LFU
	w.lfu.Set(key, value)
}

// Remove deletes an item from both LRU and LFU caches.
func (w *WindowLFU) Remove(key string) bool {
	removedFromLRU := w.lru.Remove(key)
	removedFromLFU := w.lfu.Remove(key)
	return removedFromLRU || removedFromLFU
}

// Has checks if a key exists in either LRU or LFU caches.
func (w *WindowLFU) Has(key string) bool {
	return w.lru.Has(key) || w.lfu.Has(key)
}

// GetAll returns all items from both LRU and LFU caches.
func (w *WindowLFU) GetAll(includeExpired bool) map[string]any {
	allItems := w.lru.GetAll(includeExpired)
	for key, value := range w.lfu.GetAll(includeExpired) {
		allItems[key] = value
	}
	return allItems
}

// Keys returns all keys from both LRU and LFU caches.
func (w *WindowLFU) Keys(includeExpired bool) []string {
	keys := w.lru.Keys(includeExpired)
	keys = append(keys, w.lfu.Keys(includeExpired)...)
	return keys
}

// Len returns the total count of items in both LRU and LFU caches.
func (w *WindowLFU) Len(includeExpired bool) int {
	return w.lru.Len(includeExpired) + w.lfu.Len(includeExpired)
}

// Purge clears both LRU and LFU caches.
func (w *WindowLFU) Purge() {
	w.lru.Purge()
	w.lfu.Purge()
}

func (w *WindowLFU) HitCount() uint64 {
	return w.lru.HitCount() + w.lfu.HitCount()
}

func (w *WindowLFU) MissCount() uint64 {
	return w.lru.MissCount() + w.lfu.MissCount()
}

func (w *WindowLFU) LookupCount() uint64 {
	return w.lru.LookupCount() + w.lfu.LookupCount()
}

func (w *WindowLFU) HitRate() float64 {
	hc, mc := w.lru.HitCount(), w.lfu.MissCount()
	total := hc + mc
	if total == 0 {
		return 0.0
	}
	return float64(hc) / float64(total)
}
