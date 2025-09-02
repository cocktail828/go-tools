package runnable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	pool, _ := NewHybridPool(DefaultConfig())
	cnt := 0
	for i := 0; i < 10; i++ {
		assert.NoError(t, pool.Submit(func() { cnt++ }))
	}

	pool.Close()
	assert.Equal(t, 10, cnt)
}

func TestPoolInvalidParam(t *testing.T) {
	_, err := NewHybridPool(Config{})
	assert.ErrorIs(t, err, ErrInvalidParam)
}

func TestPoolFull(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PendingTaskNum = 1
	pool, err := NewHybridPool(cfg)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 100; i++ {
		pool.Submit(func() {})
	}
	assert.Equal(t, ErrPoolFull, pool.Submit(func() {}))
}

func TestPoolClosed(t *testing.T) {
	pool, _ := NewHybridPool(DefaultConfig())
	pool.Close()
	assert.Equal(t, ErrPoolClosed, pool.Submit(func() {}))
}
