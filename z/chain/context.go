package chain

import (
	"fmt"
	"math"
	"sync"
)

type ErrCode struct {
	Code int
	Desc string
}

type Context struct {
	User    any
	data    sync.Map
	index   int
	errcode *ErrCode
}

func (c *Context) Set(k, v any) {
	c.data.Store(k, v)
}

func (c *Context) Get(k any) (any, bool) {
	return c.data.Load(k)
}

// stop iterator immediately
func (c *Context) Abort() {
	c.index = math.MaxInt16
}

// same like 'Abort', but set errcode
func (c *Context) AbortWith(code int, format string, args ...any) {
	c.errcode = &ErrCode{code, fmt.Sprintf(format, args...)}
	c.index = math.MaxInt16
}

func (c *Context) Err() *ErrCode {
	return c.errcode
}
