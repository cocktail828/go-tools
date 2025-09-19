package nacs

import (
	"context"
)

// Service 定义服务
type Service struct {
	Cluster string
	Group   string
	Name    string
	Version string
}

// Instance 定义服务实例
type Instance struct {
	Enable   bool
	Cluster  string
	Group    string
	Healthy  bool // valid at watch and discover
	Name     string
	Version  string
	Address  string // host:port
	Metadata map[string]string
}

// Registry 注册中心接口
type Registry interface {
	// Register 注册服务实例
	Register(Instance) error

	// DeRegister 注销服务实例
	DeRegister(Instance) error

	// Discover 发现服务实例
	Discover(Service) ([]Instance, error)

	// Watch 监听服务实例的变化
	Watch(svc Service, callback func([]Instance, error)) (context.CancelFunc, error)

	Close() error
}
