package chain

import (
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

// Context defines the interface for the context passed through the chain.
type Context interface {
	Temporary
	Abort()                   // Abort the chain execution.
	AbortWithError(err error) // Abort the chain execution with a custom error.
	IsAborted() bool          // Check if the chain execution is aborted.
	Next()                    // Continue to the next handler in the chain.
}

type Temporary interface {
	Set(key string, value any)  // Set a value in the context.
	Get(key string) (any, bool) // Get a value from the context.
	Request() any               // Get the request associated with the context.
}

// Handler defines the interface for a chain handler.
type Handler interface {
	Execute(ctx Context)
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

// Handle executes the chain of handlers.
func (chain *Chain) Handle(tmp Temporary) error {
	if len(chain.handlers) == 0 {
		return ErrNoHandle
	}

	// Wrap the user-provided context with our internal logic.
	cc := &chainContext{
		Temporary: tmp,
		handlers:  chain.handlers,
	}

	for cc.index < len(cc.handlers) && cc.index < abortIndex {
		cc.handlers[cc.index].Execute(cc)
		cc.index++
	}
	return cc.Error()
}

// chainContext is the internal implementation of Context.
type chainContext struct {
	Temporary
	index    int
	handlers []Handler
	error    error
}

// Abort aborts the chain execution.
func (c *chainContext) Abort() {
	c.abortWithError(errors.Errorf("chain abort at(%v) unexpectedly", c.index))
}

// AbortWithError aborts the chain execution with a custom error.
func (c *chainContext) AbortWithError(err error) {
	c.abortWithError(errors.WithMessagef(err, "chain abort at(%v) unexpectedly", c.index))
}

// abortWithError is a helper function to set the error and abort the chain.
func (c *chainContext) abortWithError(err error) {
	c.index = abortIndex
	c.error = err
}

// IsAborted checks if the chain execution is aborted.
func (c *chainContext) IsAborted() bool { return c.index >= abortIndex }

// Next continues the chain execution to the next handler.
func (c *chainContext) Next() {
	c.index++
	if c.index < len(c.handlers) && c.index < abortIndex {
		c.handlers[c.index].Execute(c)
	}
}

// Error returns the error set in the context.
func (c *chainContext) Error() error { return c.error }
