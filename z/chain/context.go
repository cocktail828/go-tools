package chain

import (
	"context"
	"log/slog"
	"sync"

	"github.com/cocktail828/go-tools/z/errcode"
	"github.com/cocktail828/go-tools/z/reflectx"
	"github.com/pkg/errors"
)

type Context struct {
	context.Context
	chain   *Chain
	index   int8
	request any
	errdesc error
	data    sync.Map
}

func (c *Context) IsAborted() bool { return !reflectx.IsNil(c.errdesc) }
func (c *Context) Abort() {
	cur := c.chain.handlers[c.index]
	c.errdesc = errors.Errorf("chain abort at(%v) unexpectlly", cur.Name())
	c.index = abortIndex
}

func (c *Context) AbortWithError(err *errcode.Error) {
	c.index = abortIndex
	c.errdesc = err
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.chain.handlers)) {
		c.chain.handlers[c.index].Execute(c)
		c.index++
	}
}

func (c *Context) Logger() *slog.Logger {
	return c.chain.Logger
}

func (c *Context) SetLocal(v any) {
	if c.index < int8(len(c.chain.handlers)) {
		cur := c.chain.handlers[c.index]
		c.data.Store(cur.Name(), v)
	}
}

func (c *Context) GetLocal() (any, bool) {
	if c.index < int8(len(c.chain.handlers)) {
		cur := c.chain.handlers[c.index]
		return c.data.Load(cur.Name())
	}
	return nil, false
}

func (c *Context) Set(v any)        { c.data.Store(globalKey, v) }
func (c *Context) Get() (any, bool) { return c.data.Load(globalKey) }
func (c *Context) Error() error     { return c.errdesc }
