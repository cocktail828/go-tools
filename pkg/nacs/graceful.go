package nacs

import (
	"context"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/z/locker"
)

type Graceful struct {
	wg       sync.WaitGroup
	Postpone time.Duration
	sync.Mutex
	Register func() DeRegister
}

func (g *Graceful) Fire(ctx context.Context) {
	g.wg = sync.WaitGroup{}
	g.wg.Add(2)

	var deregister DeRegister
	ctx, cancel := context.WithCancel(ctx)
	time.AfterFunc(g.Postpone, func() {
		defer g.wg.Done()
		locker.WithLock(g, func() {
			select {
			case <-ctx.Done():
			default:
				deregister = g.Register()
			}
		})
	})

	go func() {
		defer g.wg.Done()
		<-ctx.Done()
		locker.WithLock(g, func() {
			cancel()
			if deregister != nil {
				deregister(context.TODO())
			}
		})
	}()
}

func (g *Graceful) Wait(ctx context.Context) {
	g.wg.Wait()
}
