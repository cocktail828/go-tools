package runnable

import (
	"context"

	"github.com/pkg/errors"
)

var errNop = errors.New("nop")

type Graceful struct {
	Start func(context.Context) error // cannot be nil
	Stop  func()                      // stop should always success
}

func (g *Graceful) GoContext(inCtx context.Context) error {
	if g.Stop == nil {
		g.Stop = func() {}
	}

	ctx, cancel := context.WithCancelCause(inCtx)
	go func() {
		if err := g.Start(ctx); err != nil {
			cancel(err)
		} else {
			cancel(errNop)
		}
	}()

	// the following will block until context canceled or g.Start return error
	<-ctx.Done()

	// explicitly call the Stop function to ensure g.Start returns when the context is canceled
	g.Stop()

	if err := context.Cause(ctx); err != errNop {
		return err
	}
	return nil
}

func (g *Graceful) Go() error {
	return g.GoContext(context.Background())
}
