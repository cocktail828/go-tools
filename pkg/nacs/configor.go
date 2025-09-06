package nacs

import "context"

type GetOpt interface{ Apply() }
type MonitorOpt interface{ Apply() }
type OnChange func(error)

type Configor interface {
	Get(opts ...GetOpt) ([]byte, error)
	Monitor(cb OnChange, opts ...MonitorOpt) (context.CancelFunc, error)
	Close() error
}
