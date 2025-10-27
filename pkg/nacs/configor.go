package nacs

import "context"

type Configor interface {
	// Load load the config from config server
	Load() ([]byte, error)

	// Monitor monitor the config change
	Monitor(cb func(name string, payload []byte, err error)) (context.CancelFunc, error)
	Close() error
}
