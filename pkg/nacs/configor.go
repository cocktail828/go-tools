package nacs

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/cocktail828/go-tools/pkg/nacs/configor"
	"github.com/cocktail828/go-tools/z/environs"
)

var (
	ErrNoSuchConfigor = errors.New("no such configor")
)

type Configor struct {
	files  []string
	cache  sync.Map
	dumper dumper
	cfgor  configor.Configor
}

func NewConfigor(addr, group, service string, files ...string) (*Configor, error) {
	loader := &Configor{
		files:  files,
		dumper: dumper{".findercache", "config_"},
		cfgor:  configor.NewFileConfigor(),
	}

	if !environs.Has("NATIVE_CONFIGOR") {
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
		loader.onEvent(configor.ADD, configor.Config{k, v})
	}
	return loader, nil
}

func (loader *Configor) GetRawCfg() []byte {
	if len(loader.files) != 1 {
		panic("not exactly one config is provided")
	}

	val, ok := loader.cache.Load(loader.files[0])
	if !ok {
		return nil
	}
	return val.([]byte)
}

func (loader *Configor) GetByName(name string) []byte {
	val, ok := loader.cache.Load(name)
	if !ok {
		return nil
	}
	return val.([]byte)
}

type configChanger struct {
	*Configor
	handler configor.Handler
}

func (c configChanger) OnChange(evt configor.Event, cfg configor.Config, err error) {
	if err == nil {
		c.onEvent(evt, cfg)
	}
	c.handler.OnChange(evt, cfg, err)
}

func (loader *Configor) onEvent(evt configor.Event, cfg configor.Config) {
	switch evt {
	case configor.ADD, configor.CHG:
		loader.cache.Store(cfg.Name, cfg.Payload)
		loader.dumper.Dump(cfg.Name, cfg.Payload)

	case configor.DEL:
		loader.cache.Delete(cfg.Name)
		loader.dumper.Remove(cfg.Name)
	}
}

func (loader *Configor) Watch(ctx context.Context, handler configor.Handler) error {
	return loader.cfgor.Watch(ctx, configChanger{loader, handler}, loader.files...)
}
