package kvstore

import (
	"context"
	"errors"

	"github.com/cocktail828/go-tools/z/variadic"
)

var (
	ErrWatcherStopped = errors.New("watcher is stopped")
	ErrNotImplement   = errors.New("the method is not implement")
)

type Result interface {
	Len() int
	Key(i int) string
	Value(i int) []byte
}

type Type string

const (
	NONE   Type = "NONE"
	PUT    Type = "PUT"
	DELETE Type = "DELETE"
)

type Event interface {
	Len() int
	Key(i int) string
	Value(i int) []byte
	Type(i int) Type // PUT, DELETE
}

// KV is the source from which config is loaded.
type KV interface {
	Set(ctx context.Context, key string, val []byte, opts ...variadic.Option) error
	Get(ctx context.Context, key string, opts ...variadic.Option) (Result, error)
	Del(ctx context.Context, key string, opts ...variadic.Option) error
	Watch(ctx context.Context, opts ...variadic.Option) Watcher
	Close() error
	String() string
}

// Watcher watches a source for changes.
type Watcher interface {
	Next(ctx context.Context) (Event, error)
	Stop() error
}
