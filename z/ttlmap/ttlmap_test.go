package ttlmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpireFunc(t *testing.T) {
	f := expireFunc(time.Second)
	assert.Equal(t, true, f())
	time.Sleep(time.Second)
	assert.Equal(t, false, f())
}

func TestCache(t *testing.T) {
	c := New[string]()

	t.Run("no-ttl", func(t *testing.T) {
		c.Set("a", "val")
		val, err := c.Get("a")
		assert.Equal(t, nil, err)
		assert.Equal(t, "val", val)
		c.Del("a")
		_, err = c.Get("a")
		assert.Equal(t, ErrNoEntry, err)
	})

	t.Run("ttl", func(t *testing.T) {
		c.SetWithTTL("b", "val", time.Second)
		val, err := c.Get("b")
		assert.Equal(t, nil, err)
		assert.Equal(t, "val", val)

		time.Sleep(time.Second)
		_, err = c.Get("b")
		assert.Equal(t, ErrNoEntry, err)

		c.Del("b")
		_, err = c.Get("b")
		assert.Equal(t, ErrNoEntry, err)
	})
}
