package configuration

import "context"

type Config struct {
	Cluster  string
	Group    string
	ID       string
	Metadata map[string]string
}

// Listener 定义配置变更监听器
type Listener func(cfg Config, payload []byte, err error)

// Configor 配置中心接口
type Configor interface {
	// Get 获取配置
	Get(Config) ([]byte, error)

	// Set 设置配置
	Set(Config, []byte) error

	// Delete 删除配置
	Delete(Config) error

	// Monitor 监听配置的变化
	Monitor(Config, Listener) (context.CancelFunc, error)

	Close() error
}
