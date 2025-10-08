package runnable

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cocktail828/go-tools/z/semaphore"
	"github.com/pkg/errors"
)

type Task func()

var (
	ErrPoolClosed   = errors.New("hybrid pool: already been closed")
	ErrPoolFull     = errors.New("hybrid pool: task queue is full")
	ErrInvalidParam = errors.New("hybrid pool: invalid parameters")
)

type Config struct {
	MaxWorkers,
	MinWorkers, // default MaxWorkers / 2
	PendingTaskNum, // task queue length, default 1024
	ExpandThreshold, // expand workers if task queue length is larger than expand threshold, default min(PendingTaskNum*2/3, MaxWorkers*10)
	ShrinkThreshold int // shrink workers if task queue length is smaller than shrink threshold, default minWorkers
}

func (c *Config) Normalize() error {
	if c.MaxWorkers < c.MinWorkers || c.MinWorkers < 0 || c.MaxWorkers == 0 {
		return errors.Wrapf(ErrInvalidParam, "check parameters maxWorkers(%d), minWorkers(%d)", c.MaxWorkers, c.MinWorkers)
	}
	if c.PendingTaskNum <= 0 {
		return errors.Wrapf(ErrInvalidParam, "check parameters pendingTaskNum(%d)", c.PendingTaskNum)
	}
	if c.ShrinkThreshold <= 0 {
		return errors.Wrapf(ErrInvalidParam, "check parameters shrinkThreshold(%d)", c.ShrinkThreshold)
	}
	if c.ExpandThreshold <= 0 {
		return errors.Wrapf(ErrInvalidParam, "check parameters expandThreshold(%d)", c.ExpandThreshold)
	}
	return nil
}

func DefaultConfig() Config {
	c := Config{
		MaxWorkers:     10,
		MinWorkers:     5,
		PendingTaskNum: 1024,
	}
	c.ExpandThreshold = min(c.PendingTaskNum*2/3, c.MaxWorkers*10)
	c.ShrinkThreshold = c.MinWorkers
	return c
}

type HybridPool struct {
	c        Config
	sema     *semaphore.Weighted
	taskCh   chan Task
	shrinkCh chan struct{}
	closed   atomic.Bool
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewHybridPool(c Config) (*HybridPool, error) {
	if err := c.Normalize(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &HybridPool{
		sema:     semaphore.NewWeighted(int64(c.MaxWorkers)),
		taskCh:   make(chan Task, c.PendingTaskNum),
		shrinkCh: make(chan struct{}, c.MaxWorkers),
		ctx:      ctx,
		cancel:   cancel,
	}

	p.sema.Acquire(context.Background(), int64(c.MinWorkers))
	for i := 0; i < c.MinWorkers; i++ {
		p.wg.Add(1)
		go p.spawn()
	}

	// 1. shrink workers if task queue length is smaller than shrink threshold
	// and the number of assigned workers is larger than minWorkers
	// 2. expand workers if task queue length is larger than expand threshold
	// and the number of assigned workers is smaller than maxWorkers
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if len(p.taskCh) <= p.c.ShrinkThreshold && p.sema.Assigned() >= int64(p.c.MinWorkers) {
					p.shrinkCh <- struct{}{}
				} else if len(p.taskCh) >= c.ExpandThreshold && p.sema.TryAcquire(1) {
					p.wg.Add(1)
					go p.spawn()
				}
			case <-p.ctx.Done():
				return
			}
		}
	}()
	return p, nil
}

func (p *HybridPool) spawn() {
	defer p.wg.Done()
	defer p.sema.Release(1)

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.shrinkCh: // elastic worker stoped for idle
			return
		case task, ok := <-p.taskCh:
			if !ok {
				return
			}
			p.execute(task)
		}
	}
}

func (p *HybridPool) execute(t Task) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Task panic: %v\n", err)
		}
	}()
	t()
}

func (p *HybridPool) Close() {
	if !p.closed.CompareAndSwap(false, true) {
		return
	}

	p.cancel()
	p.wg.Wait()
	for {
		select {
		case t := <-p.taskCh:
			p.execute(t)
		default:
			return
		}
	}
}

func (p *HybridPool) Submit(task Task) (err error) {
	if p.closed.Load() {
		return ErrPoolClosed
	}

	select {
	case p.taskCh <- task:
		return nil
	default:
		return ErrPoolFull
	}
}
