package nacs

import "context"

type Config struct {
	Fname    string
	Group    string
	Metadata map[string]string
	Extra    any // 私有数据
}

// ConfigListener 定义配置变更监听器
type ConfigListener func(cfg Config, payload []byte, err error)

// Configor 配置中心接口
type Configor interface {
	// GetConfig 获取配置
	GetConfig(Config) ([]byte, error)

	// SetConfig 设置配置
	SetConfig(Config, []byte) error

	// DeleteConfig 删除配置
	DeleteConfig(Config) error

	// WatchConfig 监听配置的变化
	WatchConfig(Config, ConfigListener) (context.CancelFunc, error)

	Close() error
}
