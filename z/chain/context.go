package chain

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

type Option func(*Context)

// unmarshal and set request
func WithRequest(req any) Option {
	return func(c *Context) {
		c.data.Store(requestKey, req)
	}
}

func WithContext(ctx context.Context) Option {
	return func(c *Context) {
		c.Context = ctx
	}
}

type Context struct {
	context.Context
	chain *Chain
	index int8
	error error
	data  sync.Map
}

func (c *Context) IsAborted() bool { return c.index >= abortIndex }

func (c *Context) Abort() {
	cur := c.chain.handlers[c.index]
	c.AbortWithError(errors.Errorf("chain abort at(%v) unexpectlly", cur.Name()))
}

func (c *Context) AbortWithError(err error) {
	c.index = abortIndex
	c.error = err
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

func (c *Context) Set(v any) {
	if c.index < int8(len(c.chain.handlers)) {
		cur := c.chain.handlers[c.index]
		c.data.Store(cur.Name(), v)
	}
}

func (c *Context) Get() (any, bool) {
	if c.index < int8(len(c.chain.handlers)) {
		cur := c.chain.handlers[c.index]
		return c.data.Load(cur.Name())
	}
	return nil, false
}

func (c *Context) Error() error { return c.error }

// set instance global meta
func (c *Context) Store(v any) { c.chain.Store(v) }

// get instance global meta
func (c *Context) Load() (any, bool) { return c.chain.Load() }
