package cache

import (
	"context"
	"time"

	"github.com/cocktail828/go-tools/pkg/kvstore"
)

type readthroughOption struct{}

type Getter func(key string) ([]byte, error)

func WithReadThrough(v Getter, setopt ...kvstore.Option) kvstore.Option {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, readthroughOption{}, v)
		for _, o := range setopt {
			ctx = o(ctx)
		}
		return ctx
	}
}

type validateOption struct{}

func WithValidate(v ValidateFunc) kvstore.Option {
	return func(ctx context.Context) context.Context {
		if v != nil {
			return context.WithValue(ctx, validateOption{}, v)
		}
		return ctx
	}
}

func WithTTL(ttl time.Duration) kvstore.Option {
	return WithValidate(ExpireFunc(ttl))
}
