package chain

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/pkg/errors"
)

const (
	// abortIndex represents a typical value used in abort functions.
	abortIndex int8 = math.MaxInt8 >> 1

	// global key
	globalKey = "__global__"
)

type Handler interface {
	Name() string
	Execute(ctx *Context)
}

type Chain struct {
	Logger   *slog.Logger
	Prepare  func(context.Context) (any, error)
	handlers []Handler
}

func (chain *Chain) Use(handlers ...Handler) {
	if len(handlers) >= int(abortIndex) {
		panic("too many handlers")
	}
	chain.handlers = append(chain.handlers, handlers...)
}

func (chain *Chain) Handle(ctx context.Context) error {
	if len(chain.handlers) == 0 {
		return errors.New("no handlers found")
	}

	c := &Context{
		Context: ctx,
		chain:   chain,
	}
	defer func() {
		if err := recover(); err != nil {
			chain.Logger.Error(fmt.Sprintf("oops! chain handlers(%v) panic: %v", c.index, err))
		}
	}()

	if chain.Prepare != nil {
		req, err := chain.Prepare(ctx)
		if err != nil {
			return err
		}
		c.request = req
	}

	for _, h := range chain.handlers {
		if c.index >= int8(len(c.chain.handlers)) {
			break
		}
		h.Execute(c)
	}
	return c.Error()
}
