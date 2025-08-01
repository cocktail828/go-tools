package eviction

import (
	"sync/atomic"
	"time"
)

type Eviction interface {
	Stats
	Set(string, any)
	SetWithTTL(string, any, time.Duration)
	Get(string) (any, bool)
	Has(string) bool
	Remove(string) bool
	Keys(includeExpired bool) []string
	Len(includeExpired bool) int
	Purge()
	GetAll(includeExpired bool) map[string]any
}

type cache struct {
	size       uint
	expiration time.Duration // 0 means item has no expiration
	onEvict    func(string, any)
	onPurge    func(string, any)

	// statistics
	hitCount  uint64
	missCount uint64
}

type Stats interface {
	HitCount() uint64
	MissCount() uint64
	LookupCount() uint64
	HitRate() float64
}

// increment hit count
func (c *cache) IncrHitCount() uint64 {
	return atomic.AddUint64(&c.hitCount, 1)
}

// increment miss count
func (c *cache) IncrMissCount() uint64 {
	return atomic.AddUint64(&c.missCount, 1)
}

// HitCount returns hit count
func (c *cache) HitCount() uint64 {
	return atomic.LoadUint64(&c.hitCount)
}

// MissCount returns miss count
func (c *cache) MissCount() uint64 {
	return atomic.LoadUint64(&c.missCount)
}

// LookupCount returns lookup count
func (c *cache) LookupCount() uint64 {
	return c.HitCount() + c.MissCount()
}

// HitRate returns rate for cache hitting
func (c *cache) HitRate() float64 {
	hc, mc := c.HitCount(), c.MissCount()
	total := hc + mc
	if total == 0 {
		return 0.0
	}
	return float64(hc) / float64(total)
}
