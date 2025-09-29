package nacs

import "context"

type OnChange func(error, ...any)

type Configor interface {
	Load() ([]byte, error)
	Monitor(cb OnChange) (context.CancelFunc, error)
	Close() error
}
