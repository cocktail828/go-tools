package nacs

import "context"

// ConfigInfo contains configuration data and metadata
type ConfigInfo struct {
	DataID  string // configuration identifier (optional, implementation-specific)
	Payload []byte // configuration data
}

type Configor interface {
	// Load loads the config from config server
	Load(ctx context.Context) (ConfigInfo, error)

	// Monitor monitors the config change
	// The callback is invoked when config changes
	Monitor(cb func(info ConfigInfo, err error)) (context.CancelFunc, error)
	Close() error
}
