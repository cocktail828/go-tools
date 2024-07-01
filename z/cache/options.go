package cache

import (
	"context"
)

type OpOption struct {
	context context.Context
}

func NewOpOption(opts ...Option) *OpOption {
	oo := &OpOption{context: context.Background()}
	for _, f := range opts {
		f(oo)
	}
	return oo
}

type Option func(*OpOption)

type validateOption struct{}

func WithValidate(v ValidateFunc) Option {
	return func(oo *OpOption) {
		if v != nil {
			oo.context = context.WithValue(oo.context, validateOption{}, v)
		}
	}
}

type evictOption struct{}

func WithEvict() Option {
	return func(oo *OpOption) {
		oo.context = context.WithValue(oo.context, evictOption{}, true)
	}
}
