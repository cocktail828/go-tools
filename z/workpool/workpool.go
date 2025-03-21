package workpool

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cocktail828/go-tools/z"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

var ErrFull = errors.New("pool is full")

const (
	pendingTaskNum = 1000
)

type Task func()

type HybridPool struct {
	sema     *semaphore.Weighted
	taskChan chan Task
	ticker   *time.Ticker
	wg       sync.WaitGroup
	closed   atomic.Bool
	mu       sync.RWMutex
}

func NewHybridPool(minWorkers, maxWorkers int) (*HybridPool, error) {
	if maxWorkers < minWorkers || minWorkers < 0 || maxWorkers == 0 {
		return nil, errors.Errorf("invalid parameters, maxWorkers(%d) >= minWorkers(%d) >= 0", maxWorkers, minWorkers)
	}

	pool := &HybridPool{
		sema:     semaphore.NewWeighted(int64(maxWorkers)),
		taskChan: make(chan Task, pendingTaskNum),
		ticker:   time.NewTicker(time.Second * 10),
	}

	pool.sema.Acquire(context.Background(), int64(minWorkers))
	for i := 0; i < minWorkers; i++ {
		pool.wg.Add(1)
		go pool.spawn(false)
	}
	return pool, nil
}

func (p *HybridPool) spawn(isElastic bool) {
	defer p.wg.Done()
	defer p.sema.Release(1)

	timer := time.NewTimer(time.Minute)
	defer timer.Stop()
	for {
		select {
		case <-p.ticker.C:
			if len(p.taskChan) >= (pendingTaskNum)*3/4 {
				if p.sema.TryAcquire(1) {
					p.wg.Add(1)
					go p.spawn(true)
				}
			}

		case <-timer.C: // elastic worker stoped for idle
			if isElastic {
				return
			}

		case task, ok := <-p.taskChan:
			if !ok {
				return
			}
			task()
			timer.Reset(time.Minute)
		}
	}
}

func (p *HybridPool) Close() {
	if p.closed.CompareAndSwap(false, true) {
		p.ticker.Stop()
		z.WithLock(&p.mu, func() {
			close(p.taskChan)
		})
	}
}

func (p *HybridPool) Wait() {
	p.wg.Wait()
}

func (p *HybridPool) Submit(task Task) (err error) {
	z.WithRLock(&p.mu, func() {
		if p.closed.Load() {
			err = io.ErrClosedPipe
			return
		}

		select {
		case p.taskChan <- task:
		default:
			err = ErrFull
		}
	})
	return
}
