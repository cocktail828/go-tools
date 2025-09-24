package nacs

import (
	"context"
)

// Service 定义服务
type Service struct {
	Group string
	Name  string // service@version
}

type Instance struct {
	Enable   bool
	Group    string
	Healthy  bool   // valid at watch and discover
	Name     string // service@version
	Address  string // host:port
	Metadata map[string]string
}

type RegisterInstance struct {
	Group    string
	Name     string // service@version
	Address  string // host:port
	Metadata map[string]string
}

type DeRegisterInstance struct {
	Group   string
	Name    string // service@version
	Address string // host:port
}

// Registry 注册中心接口
type Registry interface {
	// Register 注册服务实例
	Register(RegisterInstance) (context.CancelFunc, error)

	// DeRegister 注销服务实例
	DeRegister(DeRegisterInstance) error

	// Discover 发现服务实例
	Discover(Service) ([]Instance, error)

	// Watch 监听服务实例的变化
	Watch(svc Service, callback func([]Instance, error)) (context.CancelFunc, error)

	Close() error
}
