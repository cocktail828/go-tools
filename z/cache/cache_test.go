package cache_test

import (
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/cache"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	t.Run("no-ttl", func(t *testing.T) {
		c := cache.New(cache.Config[string]{})
		c.Set("a", "val")
		s, err := c.Get("a")
		assert.Equal(t, "val", s)
		assert.Equal(t, nil, err)

		c.Del("a")
		_, err = c.Get("a")
		assert.NotEqual(t, nil, err)
	})

	t.Run("with-ttl", func(t *testing.T) {
		c := cache.New(cache.Config[string]{})
		c.Set("a", "val", cache.WithValidate(cache.ExpireFunc(time.Second)))
		s, err := c.Get("a")
		assert.Equal(t, "val", s)
		assert.Equal(t, nil, err)

		time.Sleep(time.Second)
		_, err = c.Get("a")
		assert.NotEqual(t, nil, err)
	})
}
