package chain

import (
	"math"

	"github.com/pkg/errors"
)

const (
	// abortIndex represents a typical value used in abort functions.
	abortIndex = math.MaxInt16 >> 1
)

var (
	ErrTooManyHandle = errors.Errorf("too many handlers, at most '%d' handlers is allowed", abortIndex)
)

type Object interface {
	Set(key string, value any)  // Set a value in the context.
	Get(key string) (any, bool) // Get a value from the context.
	Meta() any                  // Get the meta associated with the context.
}

// Handler defines the interface for a chain handler.
type Handler interface {
	Handle(*Context)
}

// Chain represents a chain of handlers.
type Chain struct {
	handlers []Handler
}

// Use adds handlers to the chain.
func (chain *Chain) Use(handlers ...Handler) error {
	if len(handlers)+len(chain.handlers) >= int(abortIndex) {
		return ErrTooManyHandle
	}
	chain.handlers = append(chain.handlers, handlers...)
	return nil
}

// Serve executes the chain of handlers.
func (chain *Chain) Serve(tmp Object) error {
	if len(chain.handlers) == 0 {
		return nil
	}

	// Wrap the user-provided context with our internal logic.
	cc := &Context{
		Object:   tmp,
		handlers: chain.handlers,
	}

	for cc.index < len(cc.handlers) && cc.index < abortIndex {
		cc.handlers[cc.index].Handle(cc)
		cc.index++
	}
	return cc.Error()
}

// Context is the internal implementation of Context.
type Context struct {
	Object
	index    int
	handlers []Handler
	error    error
}

// Abort aborts the chain execution.
func (c *Context) Abort() {
	c.abortWithError(errors.Errorf("chain abort at(%v) unexpectedly", c.index))
}

// AbortWithError aborts the chain execution with a custom error.
func (c *Context) AbortWithError(err error) {
	c.abortWithError(errors.WithMessagef(err, "chain abort at(%v) unexpectedly", c.index))
}

// abortWithError is a helper function to set the error and abort the chain.
func (c *Context) abortWithError(err error) {
	c.index = abortIndex
	c.error = err
}

// IsAborted checks if the chain execution is aborted.
func (c *Context) IsAborted() bool { return c.index >= abortIndex }

// Next continues the chain execution to the next handler.
func (c *Context) Next() {
	c.index++
	if c.index < len(c.handlers) && c.index < abortIndex {
		c.handlers[c.index].Handle(c)
	}
}

// Error returns the error set in the context.
func (c *Context) Error() error { return c.error }
