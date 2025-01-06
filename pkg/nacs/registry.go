package nacs

// ServiceInstance 定义服务实例
type ServiceInstance struct {
	ID       string            // 服务实例ID
	Name     string            // 服务名称
	Address  string            // 服务地址
	Port     int               // 服务端口
	Metadata map[string]string // 元数据
}

// Registry 注册中心接口
type Registry interface {
	// Register 注册服务实例
	Register(instance *ServiceInstance) error

	// Deregister 注销服务实例
	Deregister(instanceID string) error

	// Discover 发现服务实例
	Discover(serviceName string) ([]*ServiceInstance, error)

	// Watch 监听服务实例的变化
	Watch(serviceName string, callback func([]*ServiceInstance)) error
}
