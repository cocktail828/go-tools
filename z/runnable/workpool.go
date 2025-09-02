package runnable

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cocktail828/go-tools/pkg/semaphore"
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
	ExpandThreshold, // expand workers if task queue length is larger than expand threshold, default maxWorkers * 3
	ShrinkThreshold int // shrink workers if task queue length is smaller than shrink threshold, default minWorkers
	ShrinkInterval time.Duration // shrink workers interval
}

func (c *Config) Normalize() error {
	if c.MaxWorkers < c.MinWorkers || c.MinWorkers < 0 || c.MaxWorkers == 0 {
		return errors.Wrapf(ErrInvalidParam, "check parameters maxWorkers(%d), minWorkers(%d)", c.MaxWorkers, c.MinWorkers)
	}
	if c.PendingTaskNum <= 0 {
		return errors.Wrapf(ErrInvalidParam, "check parameters pendingTaskNum(%d)", c.PendingTaskNum)
	}
	if c.ShrinkInterval <= 0 {
		return errors.Wrapf(ErrInvalidParam, "check parameters shrinkInterval(%v)", c.ShrinkInterval)
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
	c.ExpandThreshold = c.MaxWorkers * 3
	c.ShrinkThreshold = c.MinWorkers
	c.ShrinkInterval = time.Second * 5
	return c
}

type HybridPool struct {
	c        Config
	sema     *semaphore.Weighted
	taskChan chan Task
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
		taskChan: make(chan Task, c.PendingTaskNum),
		ctx:      ctx,
		cancel:   cancel,
	}

	p.sema.Acquire(context.Background(), int64(c.MinWorkers))
	for i := 0; i < c.MinWorkers; i++ {
		p.wg.Add(1)
		go p.spawn(false)
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if len(p.taskChan) >= c.ExpandThreshold && p.sema.TryAcquire(1) {
					p.wg.Add(1)
					go p.spawn(true)
				}
			case <-p.ctx.Done():
				return
			}
		}
	}()
	return p, nil
}

func (p *HybridPool) spawn(isElastic bool) {
	defer p.wg.Done()
	defer p.sema.Release(1)

	timer := time.NewTimer(p.c.ShrinkInterval)
	defer timer.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-timer.C: // elastic worker stoped for idle
			if isElastic && len(p.taskChan) >= p.c.ShrinkThreshold && p.sema.Assign() <= 3 {
				return
			}
		case task, ok := <-p.taskChan:
			if !ok {
				return
			}
			p.execute(task)
			timer.Reset(p.c.ShrinkInterval)
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
		case t := <-p.taskChan:
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
	case p.taskChan <- task:
		return nil
	default:
		return ErrPoolFull
	}
}
