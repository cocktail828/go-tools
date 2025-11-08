package runnable

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	t.Parallel()
	pool, _ := NewElasticJob(DefaultConfig())
	cnt := 0
	for i := 0; i < 10; i++ {
		assert.NoError(t, pool.Submit(func() { cnt++ }))
	}

	runtime.Gosched()
	time.Sleep(100 * time.Millisecond)
	pool.Close(context.Background())
	assert.Equal(t, 0, len(pool.taskCh))
}

func TestPoolInvalidParam(t *testing.T) {
	t.Parallel()
	_, err := NewElasticJob(Config{})
	assert.ErrorIs(t, err, ErrInvalidParam)
}

func TestPoolFull(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.PendingTaskNum = 1
	pool, err := NewElasticJob(cfg)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 100; i++ {
		pool.Submit(func() {})
	}
	assert.Equal(t, ErrPoolFull, pool.Submit(func() {}))
}

func TestPoolClosed(t *testing.T) {
	t.Parallel()
	pool, _ := NewElasticJob(DefaultConfig())
	pool.Close(context.Background())
	assert.Equal(t, ErrPoolClosed, pool.Submit(func() {}))
}

func TestPoolCloseFail(t *testing.T) {
	t.Parallel()
	cnt := 102400
	c, f := context.WithCancel(context.Background())
	pool := ElasticJob{
		runningCtx:    c,
		runningCancel: f,
		taskCh:        make(chan Task, cnt),
	}
	for i := 0; i < cnt; i++ {
		pool.taskCh <- func() {}
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err := pool.Close(canceledCtx)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Greater(t, len(pool.taskCh), 0)
}

func TestPoolElastic(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.PendingTaskNum = 100
	cfg.MaxWorkers = 100
	cfg.Period = time.Second
	pool, _ := NewElasticJob(cfg)

	q := make(chan struct{}, 1)
	time.AfterFunc(time.Second*10, func() { q <- struct{}{} })
loop:
	for {
		select {
		case <-q:
			break loop
		default:
			pool.Submit(func() { time.Sleep(time.Second) })
		}
	}
}
