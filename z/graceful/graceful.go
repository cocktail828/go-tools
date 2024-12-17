package graceful

import (
	"context"
	"errors"
)

type Graceful struct {
	Start func() error
	Stop  func() error
}

func (g *Graceful) Do(ctx context.Context) error {
	runningCtx, cancel := context.WithCancelCause(ctx)
	go func() {
		cancel(g.Start())
	}()

	<-runningCtx.Done()
	return errors.Join(g.Stop(), context.Cause(runningCtx))
}
