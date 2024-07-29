package nacs

import (
	"context"
	"sync"
	"time"
)

type Graceful struct {
	wg           sync.WaitGroup
	mu           sync.Mutex
	PostponeTime time.Duration
	Register     func() DeRegister
}

func (g *Graceful) Fire(pctx context.Context) {
	g.wg = sync.WaitGroup{}
	g.wg.Add(2)

	var deregister DeRegister
	ctx, cancel := context.WithCancel(pctx)
	time.AfterFunc(g.PostponeTime, func() {
		defer g.wg.Done()

		g.mu.Lock()
		defer g.mu.Unlock()
		select {
		case <-ctx.Done():
		default:
			deregister = g.Register()
		}
	})

	go func() {
		defer g.wg.Done()

		<-ctx.Done()
		g.mu.Lock()
		defer g.mu.Unlock()
		cancel()
		if deregister != nil {
			deregister(context.Background())
		}
	}()
}

func (g *Graceful) Wait(pctx context.Context) {
	g.wg.Wait()
}
