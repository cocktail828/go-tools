package workpool_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/workpool"
	"github.com/stretchr/testify/assert"
)

func noop() { time.Sleep(100 * time.Millisecond) }

func TestPool(t *testing.T) {
	pool := workpool.NewHybridPool(3, 5)
	defer pool.Close()

	for i := 0; i < 10; i++ {
		assert.NoError(t, pool.Submit(context.Background(), noop))
	}

	pool.Close()
	pool.Wait()

	// 尝试在关闭后提交任务
	assert.Error(t, io.ErrClosedPipe, pool.Submit(context.Background(), noop))
}
