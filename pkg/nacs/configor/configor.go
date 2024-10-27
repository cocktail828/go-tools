package configor

import (
	"context"
	"sync"
)

var (
	configorMap = map[string]Builder{}
	configorMu  = sync.Mutex{}
)

func Register(n string, b Builder) {
	configorMu.Lock()
	defer configorMu.Unlock()
	configorMap[n] = b
}

func Lookup(n string) Builder {
	configorMu.Lock()
	defer configorMu.Unlock()
	return configorMap[n]
}

type Builder interface {
	Build() Configor
}

type ConfigorHandler interface {
	OnChange(Event, string, []byte, error)
}

type Configor interface {
	Watch(context.Context, ConfigorHandler, ...string) error
	Load(context.Context, ...string) (map[string][]byte, error)
}

type Event string

const (
	ADD Event = "add"
	DEL Event = "del"
	CHG Event = "chg"
	SYS Event = "SYS"
)
