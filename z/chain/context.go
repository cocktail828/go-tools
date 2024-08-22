package chain

import (
	"math"
	"sync"

	"github.com/cocktail828/go-tools/z/errcode"
)

type Context struct {
	data    sync.Map
	index   int
	errdesc errcode.Error
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
func (c *Context) AbortWith(code errcode.Code, format string, args ...any) {
	c.errdesc = errcode.Errorf(code, format, args...)
	c.index = math.MaxInt16
}

func (c *Context) Error() error {
	return c.errdesc
}
