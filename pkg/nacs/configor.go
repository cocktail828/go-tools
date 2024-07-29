package nacs

import "context"

type Configs []Config

func (cs Configs) Get(name string) []byte {
	for _, c := range cs {
		if c.Name == name {
			return c.Payload
		}
	}
	return nil
}

type Config struct {
	Name    string // 配置名称
	Payload []byte // 配置内容
}

// TODO: implement
type ConfigHandler interface {
	OnChange(event Event, cfg Config)
	ReportError(err error)
}

type Configor interface {
	LookupConfig(ctx context.Context, name string) (Config, error)
	WatchConfig(ctx context.Context, handler ConfigHandler, names ...string) error
}
