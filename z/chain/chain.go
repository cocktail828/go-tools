package chain

import (
	"context"
	"math"

	"github.com/pkg/errors"
)

var (
	ErrTooManyHandle = errors.New("too many handlers, at most 63 handlers is allowed")
	ErrNoHandle      = errors.New("chain has no handlers, forget call Use()?")
)

const (
	// abortIndex represents a typical value used in abort functions.
	abortIndex int8 = math.MaxInt8 >> 1
)

type Handler interface {
	Name() string
	Execute(ctx *Context)
}

type Chain struct {
	handlers []Handler
}

func (chain *Chain) Use(handlers ...Handler) error {
	if len(handlers)+len(chain.handlers) >= int(abortIndex) {
		return ErrTooManyHandle
	}
	chain.handlers = append(chain.handlers, handlers...)
	return nil
}

func (chain *Chain) Handle(ctx context.Context, req any) error {
	if len(chain.handlers) == 0 {
		return ErrNoHandle
	}

	c := Context{
		Context:  context.Background(),
		handlers: chain.handlers,
		req:      req,
	}

	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index].Execute(&c)
		c.index++
	}
	return c.Error()
}

type Context struct {
	context.Context
	index    int8
	handlers []Handler
	req      any
	resp     any
	private  map[string]any
	error    error
}

func (c *Context) Error() error    { return c.error }
func (c *Context) IsAborted() bool { return c.index >= abortIndex }
func (c *Context) Abort() {
	c.index = abortIndex
	cur := c.handlers[c.index]
	c.error = errors.Errorf("chain abort at(%v) unexpectlly", cur.Name())
}

func (c *Context) AbortWithError(err error) {
	c.index = abortIndex
	cur := c.handlers[c.index]
	c.error = errors.WithMessagef(err, "chain abort at(%v) unexpectlly", cur.Name())
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index].Execute(c)
		c.index++
	}
}

func (c *Context) Set(v any) {
	if c.index < int8(len(c.handlers)) {
		cur := c.handlers[c.index]
		c.private[cur.Name()] = v
	}
}

func (c *Context) Get() (any, bool) {
	if c.index < int8(len(c.handlers)) {
		cur := c.handlers[c.index]
		val, ok := c.private[cur.Name()]
		return val, ok
	}
	return nil, false
}

func (c *Context) Request() any { return c.req }
func (c *Context) Write(v any)  { c.resp = v }
func (c *Context) Result() any  { return c.resp }
