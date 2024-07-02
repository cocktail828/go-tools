package kvstore

import (
	"context"
	"errors"
)

var (
	// ErrWatcherStopped is returned when source watcher has been stopped.
	ErrWatcherStopped = errors.New("watcher stopped")
)

type KVPair struct {
	Key string
	Val []byte
}

// KV is the source from which config is loaded.
type KV interface {
	Write(ctx context.Context, key string, val []byte, opts ...Option) error
	Read(ctx context.Context, key string, opts ...Option) ([]KVPair, error)
	Delete(ctx context.Context, key string, opts ...Option) error
	Watch(ctx context.Context, key string, opts ...Option) Watcher
	Close() error
	String() string
}

//go:generate stringer -type EventType -linecomment
type EventType int

const (
	Put EventType = iota // PUT
	Del                  // DEL
)

type Event struct {
	Type EventType
	Key  string
	Val  []byte
}

// Watcher watches a source for changes.
type Watcher interface {
	Next() ([]Event, error)
	Stop() error
}
