package eviction_test

import (
	"testing"
	"time"

	"github.com/cocktail828/go-tools/algo/eviction"
	"github.com/stretchr/testify/assert"
)

func TestCacheExpiration(t *testing.T) {
	tests := []struct {
		name      string
		cacheInit func(size uint) eviction.Eviction
	}{
		{"LFU", func(size uint) eviction.Eviction { return eviction.NewLFU(size) }},
		{"LRU", func(size uint) eviction.Eviction { return eviction.NewLRU(size) }},
		{"WindowLFU", func(size uint) eviction.Eviction { return eviction.NewWindowLFU(size, size*2, 100) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.cacheInit(2)

			cache.SetWithExpiration("key1", "value1", 50*time.Millisecond)
			cache.SetWithExpiration("key2", "value2", 100*time.Millisecond)

			// 检查在过期前是否存在
			value, exists := cache.Get("key1")
			assert.True(t, exists)
			assert.Equal(t, "value1", value)

			// 等待元素过期
			time.Sleep(60 * time.Millisecond)
			_, exists = cache.Get("key1")
			assert.False(t, exists, "key1 should have expired")

			// 确认其他未过期元素仍在
			value, exists = cache.Get("key2")
			assert.True(t, exists)
			assert.Equal(t, "value2", value)
		})
	}
}

func TestCacheEviction(t *testing.T) {
	tests := []struct {
		name      string
		cacheInit func(size uint) eviction.Eviction
	}{
		{"LFU", func(size uint) eviction.Eviction { return eviction.NewLFU(size) }},
		{"LRU", func(size uint) eviction.Eviction { return eviction.NewLRU(size) }},
		{"WindowLFU", func(size uint) eviction.Eviction { return eviction.NewWindowLFU(size, size*2, 100) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.cacheInit(2)

			cache.Set("key1", "value1")
			cache.Set("key2", "value2")
			cache.Set("key3", "value3") // "key1" 应该被淘汰

			_, exists := cache.Get("key1")
			assert.False(t, exists, "key1 should have been evicted")

			cache.Get("key2")           // 提升 "key2" 的频率 (LFU) 或最近使用状态 (LRU)
			cache.Set("key4", "value4") // 淘汰其他项目 ("key3" 应该被淘汰)

			_, exists = cache.Get("key3")
			assert.False(t, exists, "key3 should have been evicted")
			value, exists := cache.Get("key2")
			assert.True(t, exists)
			assert.Equal(t, "value2", value)
		})
	}
}

func TestCacheInterfaces(t *testing.T) {
	tests := []struct {
		name      string
		cacheInit func(size uint) eviction.Eviction
	}{
		{"LFU", func(size uint) eviction.Eviction { return eviction.NewLFU(size) }},
		{"LRU", func(size uint) eviction.Eviction { return eviction.NewLRU(size) }},
		{"WindowLFU", func(size uint) eviction.Eviction { return eviction.NewWindowLFU(size, size*2, 100) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.cacheInit(3)

			cache.Set("key1", "value1")
			cache.Set("key2", "value2")
			cache.Set("key3", "value3")

			assert.True(t, cache.Has("key1"))
			assert.True(t, cache.Has("key2"))
			assert.False(t, cache.Has("nonexistent"))

			cache.Remove("key2")
			assert.False(t, cache.Has("key2"), "key2 should have been removed")

			allItems := cache.GetAll(false)
			assert.Equal(t, 2, len(allItems))
			assert.Contains(t, allItems, "key1")
			assert.Contains(t, allItems, "key3")
		})
	}
}

func TestCachePurge(t *testing.T) {
	tests := []struct {
		name      string
		cacheInit func(size uint) eviction.Eviction
	}{
		{"LFU", func(size uint) eviction.Eviction { return eviction.NewLFU(size) }},
		{"LRU", func(size uint) eviction.Eviction { return eviction.NewLRU(size) }},
		{"WindowLFU", func(size uint) eviction.Eviction { return eviction.NewWindowLFU(size, size*2, 100) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.cacheInit(2)
			cache.Set("key1", "value1")
			cache.Set("key2", "value2")

			cache.Purge()
			assert.Equal(t, 0, cache.Len(false), "Cache should be empty after purge")
		})
	}
}
