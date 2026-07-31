package nacs

import (
	"context"
)

// Instance represents a service instance
type Instance struct {
	Service string            // service name
	Version string            // service version
	Host    string            // instance host
	Port    uint              // instance port
	Meta    map[string]string // instance metadata
}

// Registry is a service registry interface
// Service details such as service name, service version, and cluster information
// are determined by the registry implementation (typically via construction parameters)
type Registry interface {
	// Register registers a service instance
	// Returns a cancel function to deregister the instance automatically
	Register(ctx context.Context, inst Instance) (context.CancelFunc, error)

	// DeRegister de-registers a service instance
	DeRegister(ctx context.Context, inst Instance) error

	// Discover discovers service instances
	// The discovery scope (namespace, service, version) is determined by the registry implementation
	Discover(ctx context.Context) ([]Instance, error)

	// Watch watches service instance changes
	// The callback is invoked when instances change
	Watch(callback func([]Instance, error)) (context.CancelFunc, error)

	Close() error
}
