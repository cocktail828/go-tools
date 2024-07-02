package kvstore

import "context"

type Options struct {
	Context context.Context
}

type Option func(*Options)

func NewOptions(opts ...Option) *Options {
	o := &Options{context.Background()}
	for _, f := range opts {
		f(o)
	}
	return o
}
