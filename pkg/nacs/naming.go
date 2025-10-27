package nacs

import (
	"context"
	"strings"
)

type Instance struct {
	Enable   bool   // valid at watch and discover
	Healthy  bool   // valid at watch and discover
	Service  string // service@version format, valid at watch and discover
	Host     string // host
	Port     uint   // port
	Metadata map[string]string
}

func Compose(service, version string) string {
	return service + "@" + version
}

func (i Instance) ServiceVersion() (string, string) {
	service, version, _ := strings.Cut(i.Service, "@")
	return service, version
}

// Registry is a service registry interface
// Service details such as service name, service version, and cluster information
// must be provided and handled in specific implementations
type Registry interface {
	// Register register a service instance
	Register(instance Instance) (context.CancelFunc, error)

	// DeRegister de-registers a service instance
	DeRegister(instance Instance) error

	// Discover discovers service instances
	Discover() ([]Instance, error)

	// Watch watches service instance changes
	Watch(callback func([]Instance, error)) (context.CancelFunc, error)

	Close() error
}
