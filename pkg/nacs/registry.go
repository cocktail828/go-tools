package nacs

// Service 定义服务实例
type Service struct {
	Cluster  string
	Group    string
	Name     string
	ID       string            // 服务实例ID
	IP       string            // 服务地址
	Port     int               // 服务端口
	Metadata map[string]string // 元数据
	Extra    any               // 私有数据
}

type ServiceFilter struct {
	Cluster string
	Group   string
	Name    string
}

// Registry 注册中心接口
type Registry interface {
	// Register 注册服务实例
	Register(svc Service) error

	// Deregister 注销服务实例
	Deregister(svc Service) error

	// Discover 发现服务实例
	Discover(sf ServiceFilter) ([]Service, error)

	// Watch 监听服务实例的变化
	Watch(sf ServiceFilter, callback func([]Service, error)) error

	Close() error
}
