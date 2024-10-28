package configor

import (
	"context"
	"strings"
	"sync"
)

var (
	configorMap = map[string]Builder{}
	configorMu  = sync.Mutex{}
)

func Register(n string, b Builder) {
	configorMu.Lock()
	defer configorMu.Unlock()
	configorMap[strings.ToLower(n)] = b
}

func Lookup(n string) Builder {
	configorMu.Lock()
	defer configorMu.Unlock()
	return configorMap[strings.ToLower(n)]
}

type Builder interface {
	Build() Configor
}

type Config struct {
	Name    string `json:"name,omitempty"`
	Payload []byte `json:"payload,omitempty"`
}

type Handler interface {
	OnChange(Event, Config, error)
}

type Configor interface {
	Watch(context.Context, Handler, ...string) error
	Load(context.Context, ...string) (map[string][]byte, error)
}

type Event string

const (
	ADD Event = "add"
	DEL Event = "del"
	CHG Event = "chg"
	SYS Event = "SYS"
)
