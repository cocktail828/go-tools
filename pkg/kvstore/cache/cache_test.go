package cache_test

import (
	"testing"
	"time"

	"github.com/cocktail828/go-tools/pkg/kvstore"
	"github.com/cocktail828/go-tools/pkg/kvstore/cache"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	t.Run("no-ttl", func(t *testing.T) {
		c := cache.New()
		c.Set("a", []byte("val"))
		s, err := c.Get("a")
		assert.Equal(t, []kvstore.KVPair{{Key: "a", Val: []byte("val")}}, s)
		assert.Equal(t, nil, err)

		c.Del("a")
		_, err = c.Get("a")
		assert.NotEqual(t, nil, err)
	})

	t.Run("with-ttl", func(t *testing.T) {
		c := cache.New()
		c.Set("a", []byte("val"), cache.WithValidate(cache.ExpireFunc(time.Second)))
		s, err := c.Get("a")
		assert.Equal(t, []kvstore.KVPair{{Key: "a", Val: []byte("val")}}, s)
		assert.Equal(t, nil, err)

		time.Sleep(time.Second)
		_, err = c.Get("a")
		assert.NotEqual(t, nil, err)
	})
}
