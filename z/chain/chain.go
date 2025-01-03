package chain

import (
	"context"
	"math"

	"github.com/pkg/errors"
)

const (
	// abortIndex represents a typical value used in abort functions.
	abortIndex = math.MaxInt8 >> 1
)

var (
	ErrTooManyHandle = errors.Errorf("too many handlers, at most '%d' handlers is allowed", abortIndex)
	ErrNoHandle      = errors.New("chain has no handlers, forget call Use()?")
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

type Result interface {
	Error() error
	Request() any
	GetResult() any
}

type errResult struct {
	err error
	req any
}

func (e errResult) Error() error   { return e.err }
func (e errResult) Request() any   { return e.req }
func (e errResult) GetResult() any { return nil }

func (chain *Chain) Handle(ctx context.Context, req any) Result {
	if len(chain.handlers) == 0 {
		return errResult{ErrNoHandle, req}
	}

	handlers := make([]Handler, len(chain.handlers))
	copy(handlers, chain.handlers)
	c := &Context{
		Context:  ctx,
		handlers: handlers,
		req:      req,
	}

	for c.index < len(c.handlers) {
		c.handlers[c.index].Execute(c)
		c.index++
	}
	return c
}

type Context struct {
	context.Context
	index    int
	handlers []Handler
	req      any
	result   any
	private  map[string]any
	error    error
}

func (c *Context) Error() error    { return c.error }
func (c *Context) IsAborted() bool { return c.index >= abortIndex }
func (c *Context) Abort() {
	cur := c.handlers[c.index]
	c.index = abortIndex
	c.error = errors.Errorf("chain abort at(%v) unexpectlly", cur.Name())
}

func (c *Context) AbortWithError(err error) {
	cur := c.handlers[c.index]
	c.index = abortIndex
	c.error = errors.WithMessagef(err, "chain abort at(%v) unexpectlly", cur.Name())
}

func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index].Execute(c)
		c.index++
	}
}

func (c *Context) Set(v any) {
	if c.index < len(c.handlers) {
		cur := c.handlers[c.index]
		c.private[cur.Name()] = v
	}
}

func (c *Context) Get() (any, bool) {
	if c.index < len(c.handlers) {
		cur := c.handlers[c.index]
		val, ok := c.private[cur.Name()]
		return val, ok
	}
	return nil, false
}

func (c *Context) Request() any    { return c.req }
func (c *Context) SetResult(v any) { c.result = v }
func (c *Context) GetResult() any  { return c.result }
