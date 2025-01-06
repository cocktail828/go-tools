package workpool_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/pkg/workpool"
	"github.com/stretchr/testify/assert"
)

type simpleTask struct {
	id int
}

func (t *simpleTask) Do() {
	// fmt.Printf("Task %d is running\n", t.id)
	time.Sleep(100 * time.Millisecond)
}

func TestPool(t *testing.T) {
	pool := workpool.NewHybridPool(3, 5)
	defer pool.Close()

	for i := 0; i < 10; i++ {
		task := &simpleTask{id: i}
		assert.NoError(t, pool.Submit(context.Background(), task))
	}

	pool.Close()
	pool.Wait()

	// 尝试在关闭后提交任务
	assert.Error(t, io.ErrClosedPipe, pool.Submit(context.Background(), &simpleTask{id: 11}))
}
