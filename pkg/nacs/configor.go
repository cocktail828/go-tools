package nacs

// ConfigListener 定义配置变更监听器
type ConfigListener func(fname string, payload []byte, err error)

// Configor 配置中心接口
type Configor interface {
	// GetConfig 获取配置项的值
	GetConfig(fname string) ([]byte, error)

	// SetConfig 设置配置项的值
	SetConfig(fname string, payload []byte) error

	// DeleteConfig 删除配置项
	DeleteConfig(fname string) error

	// WatchConfig 监听配置项的变化
	WatchConfig(listener ConfigListener) error
}
