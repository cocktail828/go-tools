package kv

import (
	"context"
	"errors"
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

type SetOption func(any)
type GetOption func(any)
type DelOption func(any)
type WatchOption func(any)

// KV is the source from which config is loaded.
type KV interface {
	Set(ctx context.Context, key string, val []byte, opts ...SetOption) error
	Get(ctx context.Context, key string, opts ...GetOption) (Result, error)
	Del(ctx context.Context, key string, opts ...DelOption) error
	Watch(ctx context.Context, opts ...WatchOption) Watcher
	Close() error
	String() string
}

type Watcher interface {
	Next(ctx context.Context) (Event, error)
	Stop() error
}
