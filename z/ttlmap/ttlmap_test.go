package ttlmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	cache := New[string]()

	// Test Set and Get
	cache.Set("key1", "value1")
	val, err := cache.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Test SetWithTTL and expiration
	cache.SetWithTTL("key2", "value2", 100*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	val, err = cache.Get("key2")
	assert.Error(t, err)
	assert.Equal(t, ErrNoEntry, err)

	// Test GetDel
	cache.Set("key3", "value3")
	val, err = cache.GetDel("key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)
	val, err = cache.Get("key3")
	assert.Error(t, err)
	assert.Equal(t, ErrNoEntry, err)

	// Test Del
	cache.Set("key4", "value4")
	oldVal, ok := cache.Del("key4")
	assert.True(t, ok)
	assert.Equal(t, "value4", oldVal)
	val, err = cache.Get("key4")
	assert.Error(t, err)
	assert.Equal(t, ErrNoEntry, err)
}
