package chain

import "log"

type Handler func(*Context)

type Chain struct {
	Handlers []Handler
}

func (c *Chain) Use(h Handler) {
	c.Handlers = append(c.Handlers, h)
}

func (c *Chain) Handle(ctx *Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("oops! chain handlers(%v) panic: %v", ctx.index, err)
		}
	}()

	for ctx.index < len(c.Handlers) {
		c.Handlers[ctx.index](ctx)
		ctx.index++
	}
}
