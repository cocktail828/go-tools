package kvstore

import (
	"errors"
)

var (
	ErrWatcherStopped = errors.New("watcher stopped")
	ErrNotImplement   = errors.New("method not implement")
)

type KVPair struct {
	Key string
	Val []byte
}

// KV is the source from which config is loaded.
type KV interface {
	Set(key string, val []byte, opts ...Option) error
	Get(key string, opts ...Option) ([]KVPair, error)
	Del(key string, opts ...Option) error
	Watch(opts ...Option) Watcher
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

type NopWatcher struct{}

func (NopWatcher) Next() ([]Event, error) { return nil, ErrNotImplement }
func (NopWatcher) Stop() error            { return nil }
