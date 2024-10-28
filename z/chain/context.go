package chain

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

type Option func(*Context)

// unmarshal and set request
func WithRequest(req any) Option {
	return func(ctx *Context) { ctx.data.Store(requestKey, req) }
}

type Context struct {
	context.Context
	chain   *Chain
	index   int8
	isAbort atomic.Bool
	errdesc error
	data    sync.Map
}

func (c *Context) IsAborted() bool { return c.isAbort.Load() }

func (c *Context) Abort() {
	cur := c.chain.handlers[c.index]
	c.AbortWithError(errors.Errorf("chain abort at(%v) unexpectlly", cur.Name()))
}

func (c *Context) AbortWithError(err error) {
	c.index = abortIndex
	c.errdesc = err
	c.isAbort.Store(true)
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.chain.handlers)) {
		c.chain.handlers[c.index].Execute(c)
		c.index++
	}
}

func (c *Context) Request() (any, bool) {
	return c.data.Load(requestKey)
}

func (c *Context) Logger() *slog.Logger {
	return c.chain.Logger
}

func (c *Context) Store(v any) {
	if c.index < int8(len(c.chain.handlers)) {
		cur := c.chain.handlers[c.index]
		c.data.Store(cur.Name(), v)
	}
}

func (c *Context) Load() (any, bool) {
	if c.index < int8(len(c.chain.handlers)) {
		cur := c.chain.handlers[c.index]
		return c.data.Load(cur.Name())
	}
	return nil, false
}

func (c *Context) Error() error { return c.errdesc }

// set instance global meta
func (c *Context) StoreMeta(v any) { c.chain.StoreMeta(v) }

// get instance global meta
func (c *Context) LoadMeta() (any, bool) { return c.chain.LoadMeta() }
