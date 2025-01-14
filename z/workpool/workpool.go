package workpool

import (
	"context"
	"io"
	"sync"
	"sync/atomic"

	"github.com/cocktail828/go-tools/z"
)

const (
	pendingTaskNum = 1000
)

type Task interface{ Do() }

type HybridPool struct {
	tickets  chan struct{} // elastic workers
	taskChan chan Task
	wg       sync.WaitGroup
	closed   atomic.Bool
	mu       sync.Mutex
}

func NewHybridPool(minWorkers, maxWorkers int) *HybridPool {
	if maxWorkers < minWorkers || minWorkers < 0 || maxWorkers == 0 {
		panic("invalid parameters")
	}

	pool := &HybridPool{
		tickets:  make(chan struct{}, maxWorkers-minWorkers),
		taskChan: make(chan Task, pendingTaskNum),
	}
	for i := 0; i < minWorkers; i++ {
		pool.wg.Add(1)
		go pool.spawn()
	}
	return pool
}

func (p *HybridPool) spawn() {
	defer p.wg.Done()

	for {
		if task, ok := <-p.taskChan; !ok {
			select { // elastic worker stoped for idle
			case <-p.tickets:
				return
			default: // closed?
				if p.closed.Load() {
					return
				}
			}
		} else {
			task.Do()
		}
	}
}

func (p *HybridPool) Close() {
	if p.closed.CompareAndSwap(false, true) {
		z.WithLock(&p.mu, func() {
			close(p.taskChan)
		})
	}
}

func (p *HybridPool) Wait() {
	p.wg.Wait()
}

func (p *HybridPool) Spawn() {
	select {
	case p.tickets <- struct{}{}:
		p.wg.Add(1)
		go p.spawn()
	default:
	}
}

func (p *HybridPool) Submit(ctx context.Context, task Task) error {
	if p.closed.Load() {
		return io.ErrClosedPipe
	}

	if len(p.taskChan) >= (pendingTaskNum)*3/4 {
		p.Spawn()
	}

	var err error
	z.WithLock(&p.mu, func() {
		select {
		case <-ctx.Done():
			err = ctx.Err()

		case p.taskChan <- task:
		}
	})
	return err
}
