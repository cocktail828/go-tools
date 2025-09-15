package nacs

import "context"

type LoadOpt func(any)
type MonitorOpt func(any)
type OnChange func(error)

type Configor interface {
	Load(opts ...LoadOpt) ([]byte, error)
	Monitor(cb OnChange, opts ...MonitorOpt) (context.CancelFunc, error)
	Close() error
}
