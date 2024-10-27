package nacs

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/cocktail828/go-tools/pkg/nacs/configor"
)

var (
	ErrNoSuchConfigor = errors.New("no such configor")
)

type ConfigLoader struct {
	files []string
	cache sync.Map
	cfgor configor.Configor
}

func NewConfigor(addr, group, service string, native bool, files ...string) (*ConfigLoader, error) {
	loader := &ConfigLoader{
		files: files,
		cfgor: configor.NewFileConfigor(),
	}

	if !native {
		urld, err := url.Parse(addr)
		if err != nil {
			return nil, err
		}

		builder := configor.Lookup(urld.Scheme)
		if builder == nil {
			return nil, ErrNoSuchConfigor
		}
		loader.cfgor = builder.Build()
	}

	vals, err := loader.cfgor.Load(context.Background(), files...)
	if err != nil {
		return nil, err
	}
	for k, v := range vals {
		loader.cache.Store(k, v)
	}
	return loader, nil
}

func (loader *ConfigLoader) GetRawCfg() []byte {
	if len(loader.files) != 1 {
		panic("not exactly one config is provided")
	}

	val, ok := loader.cache.Load(loader.files[0])
	if !ok {
		return nil
	}
	return val.([]byte)
}

func (loader *ConfigLoader) GetByName(name string) []byte {
	val, ok := loader.cache.Load(name)
	if !ok {
		return nil
	}
	return val.([]byte)
}

type changer struct {
	*ConfigLoader
	handler configor.ConfigorHandler
}

func (c changer) OnChange(evt configor.Event, name string, payload []byte, err error) {
	if err == nil {
		switch evt {
		case configor.ADD, configor.CHG:
			c.cache.Store(name, payload)
		case configor.DEL:
			c.cache.Delete(name)
		}
	}
	c.handler.OnChange(evt, name, payload, err)
}

func (loader *ConfigLoader) Watch(ctx context.Context, handler configor.ConfigorHandler) error {
	return loader.cfgor.Watch(ctx, changer{loader, handler}, loader.files...)
}
