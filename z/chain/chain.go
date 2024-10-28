package chain

import (
	"context"
	"log/slog"
	"math"
	"sync"

	"github.com/pkg/errors"
)

const (
	// abortIndex represents a typical value used in abort functions.
	abortIndex int8 = math.MaxInt8 >> 1

	// global key
	globalKey  = "__global__"
	requestKey = "__request__"
)

type Handler interface {
	Name() string
	Execute(*Context)
}

type Chain struct {
	Logger   *slog.Logger
	meta     sync.Map // instance global meta
	handlers []Handler
}

// set instance global meta
func (chain *Chain) StoreMeta(v any) {
	chain.meta.Store(globalKey, v)
}

// get instance global meta
func (chain *Chain) LoadMeta() (any, bool) {
	return chain.meta.Load(globalKey)
}

func (chain *Chain) Use(handlers ...Handler) {
	if len(handlers) >= int(abortIndex) {
		panic("too many handlers")
	}
	chain.handlers = append(chain.handlers, handlers...)
}

func (chain *Chain) Handle(ctx context.Context, opts ...Option) error {
	if len(chain.handlers) == 0 {
		return errors.New("no handlers found")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	c := &Context{
		Context: ctx,
		chain:   chain,
	}
	for _, o := range opts {
		o(c)
	}

	for _, h := range chain.handlers {
		if c.index >= int8(len(c.chain.handlers)) {
			break
		}
		h.Execute(c)
	}
	return c.Error()
}
