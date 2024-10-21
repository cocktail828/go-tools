package nacs

import (
	"context"
	"errors"
	"time"
)

type Graceful struct {
	Postpone time.Duration
	Start    func() error
	Stop     func() error
}

func (g *Graceful) Do(ctx context.Context) error {
	runningCtx, cancel := context.WithCancelCause(ctx)
	go func() {
		<-time.After(g.Postpone)
		if err := g.Start(); err != nil {
			cancel(err)
		}
	}()

	<-runningCtx.Done()
	return errors.Join(runningCtx.Err(), g.Stop())
}
