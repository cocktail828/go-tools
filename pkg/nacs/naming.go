package nacs

import (
	"context"
)

type Instance struct {
	Name string // expect service@version format, valid at watch and discover
	Host string // host
	Port uint   // port
	Meta map[string]string
}

// Registry is a service registry interface
// Service details such as service name, service version, and cluster information
// must be provided and handled in specific implementations
type Registry interface {
	// Register register a service instance
	Register(host string, port uint, meta map[string]string) (context.CancelFunc, error)

	// DeRegister de-registers a service instance
	DeRegister(host string, port uint) error

	// Discover discovers service instances
	Discover() ([]Instance, error)

	// Watch watches service instance changes
	Watch(callback func([]Instance, error)) (context.CancelFunc, error)

	Close() error
}
