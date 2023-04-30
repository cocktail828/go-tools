package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type graceful struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	signalChan chan os.Signal
}

func WithContext(ctx context.Context) *graceful {
	ctx, cancel := context.WithCancel(ctx)
	g := &graceful{
		ctx:        ctx,
		cancel:     cancel,
		signalChan: make(chan os.Signal),
	}
	signal.Notify(g.signalChan)
	signal.Ignore(syscall.SIGURG, syscall.SIGWINCH)

	return g
}

func (g *graceful) Run(runable func(context.Context)) {
	g.wg.Add(1)
	defer g.wg.Done()

	go runable(g.ctx)
}

func (g *graceful) Stop() {
	signal.Stop(g.signalChan)
	g.cancel()
}

// If f returns false, break the infinite loop.
func (g *graceful) Wait(f func(sig os.Signal) bool) {
_loop:
	for {
		select {
		case sig := <-g.signalChan:
			if f == nil {
				break _loop
			}

			if !f(sig) {
				g.cancel()
			}

		case <-g.ctx.Done():
			break _loop
		}
	}

	g.cancel()
	g.wg.Wait()
}
