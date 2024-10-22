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
	Parser   func(context.Context) (any, error)
	meta     any // instance global meta
	handlers []Handler
}

// set instance global meta
func (chain *Chain) SetMeta(v any) {
	chain.meta = v
}

// get instance global meta
func (chain *Chain) GetMeta() any {
	return chain.meta
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

	if chain.Parser != nil {
		req, err := chain.Parser(ctx)
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
