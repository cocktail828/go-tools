package etcd

import (
	"context"
	"strings"
	"time"

	"github.com/cocktail828/go-tools/pkg/kvstore"
)

type addressKey struct{}

// WithAddress sets the etcd address.
func WithAddress(a ...string) kvstore.Option {
	return func(o *kvstore.Options) {
		o.Context = context.WithValue(o.Context, addressKey{}, a)
	}
}

type prefixKey struct{}

// WithPrefix sets the key prefix to use.
func WithPrefix(p string) kvstore.Option {
	return func(o *kvstore.Options) {
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		o.Context = context.WithValue(o.Context, prefixKey{}, p)
	}
}

type authKey struct{}
type authCreds struct {
	Username string
	Password string
}

// Auth allows you to specify username/password.
func WithAuth(username, password string) kvstore.Option {
	return func(o *kvstore.Options) {
		o.Context = context.WithValue(o.Context, authKey{}, &authCreds{Username: username, Password: password})
	}
}

type dialTimeoutKey struct{}

// WithDialTimeout set the time out for dialing to etcd.
func WithDialTimeout(timeout time.Duration) kvstore.Option {
	return func(o *kvstore.Options) {
		o.Context = context.WithValue(o.Context, dialTimeoutKey{}, timeout)
	}
}

type matchPrefix struct{}

// MatchPrefix is used for read, watch, delete.
func MatchPrefix() kvstore.Option {
	return func(o *kvstore.Options) {
		o.Context = context.WithValue(o.Context, matchPrefix{}, true)
	}
}
