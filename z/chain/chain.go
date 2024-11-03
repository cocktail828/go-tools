package chain

import (
	"context"
	"math"
	"sync"

	"github.com/pkg/errors"
)

var (
	ErrTooManyHandle = errors.New("too many handlers, at most 63 handlers is allowed")
	ErrNoHandle      = errors.New("chain has no handlers, forget call Use()?")
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
	meta     sync.Map // instance global meta
	handlers []Handler
}

// set instance global meta
func (chain *Chain) Store(v any) {
	chain.meta.Store(globalKey, v)
}

// get instance global meta
func (chain *Chain) Load() (any, bool) {
	return chain.meta.Load(globalKey)
}

func (chain *Chain) Use(handlers ...Handler) error {
	if len(handlers) >= int(abortIndex) {
		return ErrTooManyHandle
	}
	chain.handlers = append(chain.handlers, handlers...)
	return nil
}

func (chain *Chain) Handle(opts ...Option) error {
	if len(chain.handlers) == 0 {
		return ErrNoHandle
	}

	c := &Context{
		Context: context.Background(),
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
