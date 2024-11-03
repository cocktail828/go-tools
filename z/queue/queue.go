package queue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrTooManyTask = errors.New("the length of queue exceed the limit")
	ErrClosed      = errors.New("the work queue already been closed")
)

type Handle func(context.Context)
type Queue struct {
	ctx         context.Context
	taskq       chan Handle
	incrq       chan struct{}
	decrq       chan struct{}
	concurrency int
	mu          sync.RWMutex
	isclosed    atomic.Bool
	wg          sync.WaitGroup
}

// default concurrency 5
func WithContext(ctx context.Context) *Queue {
	q := &Queue{
		ctx:   ctx,
		taskq: make(chan Handle, 1024),
		incrq: make(chan struct{}, 10),
		decrq: make(chan struct{}, 10),
		wg:    sync.WaitGroup{},
	}

	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		for {
			select {
			case <-q.incrq:
				q.wg.Add(1)
				go q.spawn()
			case <-q.ctx.Done():
				if q.isclosed.CompareAndSwap(false, true) {
					q.mu.Lock()
					close(q.taskq)
					q.mu.Unlock()
				}
				return
			}
		}
	}()
	q.SetConcurrency(5)
	return q
}

func (q *Queue) spawn() {
	defer q.wg.Done()
	for {
		select {
		case <-q.decrq:
			return
		case t, ok := <-q.taskq:
			if !ok {
				return
			}
			t(q.ctx)
		}
	}
}

func (q *Queue) Concurrency() int { return q.concurrency }
func (q *Queue) SetConcurrency(n int) {
	if n < 0 {
		return
	}

	for i := n; i < q.concurrency; i++ {
		q.decrq <- struct{}{}
	}
	for i := q.concurrency; i < n; i++ {
		q.incrq <- struct{}{}
	}
	q.concurrency = n
}

func (q *Queue) GoContext(ctx context.Context, t Handle) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.isclosed.Load() {
		return ErrClosed
	}

	select {
	case q.taskq <- t:
		return nil
	case <-ctx.Done():
		return context.DeadlineExceeded
	}
}

func (q *Queue) Go(t Handle) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.isclosed.Load() {
		return ErrClosed
	}

	select {
	case q.taskq <- t:
		return nil
	default:
		return ErrTooManyTask
	}
}

func (q *Queue) Wait() {
	q.wg.Wait()
}
