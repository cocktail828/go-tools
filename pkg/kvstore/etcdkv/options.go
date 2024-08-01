package etcdkv

import (
	"context"
	"strings"
	"time"

	"github.com/cocktail828/go-tools/pkg/kvstore"
)

type contextKey struct{}

// WithContext sets the etcd address.
func WithContext(ctx context.Context) kvstore.Option {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, contextKey{}, ctx)
	}
}

type addressKey struct{}

// WithAddress sets the etcd address.
func WithAddress(a ...string) kvstore.Option {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, addressKey{}, a)
	}
}

type prefixKey struct{}

// WithPrefix sets the key prefix to use.
func WithPrefix(p string) kvstore.Option {
	return func(ctx context.Context) context.Context {
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		return context.WithValue(ctx, prefixKey{}, p)
	}
}

type authKey struct{}
type authCreds struct {
	Username string
	Password string
}

// Auth allows you to specify username/password.
func WithAuth(username, password string) kvstore.Option {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, authKey{}, &authCreds{Username: username, Password: password})
	}
}

type dialTimeoutKey struct{}

// WithDialTimeout set the time out for dialing to etcd.
func WithDialTimeout(timeout time.Duration) kvstore.Option {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, dialTimeoutKey{}, timeout)
	}
}

type matchPrefix struct{}

// MatchPrefix is used for read, watch, delete.
func MatchPrefix() kvstore.Option {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, matchPrefix{}, true)
	}
}

type watchKey struct{}

// WatchKey is used for read, watch, delete.
func WatchKey(key string) kvstore.Option {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, watchKey{}, key)
	}
}
