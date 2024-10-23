package workqueue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrTooManyTask = errors.New("the length of task queue exceed the limit")
	ErrClosed      = errors.New("the work queue already been closed")
)

type Task interface {
	Handle(context.Context)
}

type Workq struct {
	runningCtx context.Context
	cancel     context.CancelFunc
	taskq      chan Task
	incq       chan struct{}
	decq       chan struct{}
	wg         sync.WaitGroup
	n          atomic.Int32
}

// default concurrency=10
func New() *Workq {
	ctx, cancel := context.WithCancel(context.Background())
	wq := &Workq{
		runningCtx: ctx,
		cancel:     cancel,
		taskq:      make(chan Task, 1024),
		incq:       make(chan struct{}, 10),
		decq:       make(chan struct{}, 10),
		wg:         sync.WaitGroup{},
	}
	wq.ExpandN(10)
	return wq
}

func (wq *Workq) Kickoff(t Task) error {
	select {
	case wq.taskq <- t:
		return nil
	case <-wq.runningCtx.Done():
		return ErrClosed
	default:
		return ErrTooManyTask
	}
}

func (wq *Workq) Stop() {
	wq.cancel()
	wq.wg.Wait()
}

func (wq *Workq) N() int {
	return int(wq.n.Load())
}

func (wq *Workq) ExpandN(n int) {
	for i := 0; i < n; i++ {
		wq.incq <- struct{}{}
	}
}

func (wq *Workq) NarrowN(n int) {
	for i := 0; i < n; i++ {
		if wq.n.Load() > 0 {
			wq.decq <- struct{}{}
		}
	}
}

func (wq *Workq) Start() {
	wq.wg.Add(1)
	go func() {
		defer wq.wg.Done()

		select {
		case <-wq.runningCtx.Done():
			return
		case <-wq.incq:
			wq.wg.Add(1)
			go wq.do()
		}
	}()
}

func (wq *Workq) do() {
	defer wq.wg.Done()
	wq.n.Add(1)
	defer wq.n.Add(-1)

	select {
	case <-wq.runningCtx.Done():
		return
	case <-wq.decq:
		return
	case t := <-wq.taskq:
		t.Handle(wq.runningCtx)
	}
}
