package nacs

import "context"

type OnChange func(name string, payload []byte, err error)

type Configor interface {
	Load() ([]byte, error)
	Monitor(cb OnChange) (context.CancelFunc, error)
	Close() error
}
