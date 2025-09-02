package variadic

import (
	"context"
)

type Container interface {
	context.Context
}

type Option func(c Container) Container

func Set(key any, value any) Option {
	return func(c Container) Container {
		return context.WithValue(c, key, value)
	}
}

func Get[T any](c Container, key any) (T, bool) {
	v, ok := c.Value(key).(T)
	if !ok {
		var zero T
		return zero, false
	}
	return v, true
}

func Value[T any](c Container, key any) T {
	v, ok := c.Value(key).(T)
	if !ok {
		var zero T
		return zero
	}
	return v
}

func Compose(opts ...Option) Container {
	c := context.Background()
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}
