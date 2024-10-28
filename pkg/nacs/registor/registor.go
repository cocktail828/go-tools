package registor

import (
	"context"
	"sync"
)

var (
	registorMap = map[string]Builder{}
	registorMu  = sync.Mutex{}
)

func Register(n string, b Builder) {
	registorMu.Lock()
	defer registorMu.Unlock()
	registorMap[n] = b
}

func Lookup(n string) Builder {
	registorMu.Lock()
	defer registorMu.Unlock()
	return registorMap[n]
}

type Builder interface {
	Build() Registor
}

type Handler interface {
	OnChange(Event, ServiceMeta, error)
}

type ServiceMeta struct {
	Service
	Meta    map[string]string
	Tags    []string
	Address string // ip:port
}

type Service struct {
	Name    string // service name
	Version string // service version
}

func (svc Service) String() string { return svc.Name + "#" + svc.Version }

type Registor interface {
	Watch(ctx context.Context, h Handler, svcs ...Service) error
	Lookup(svc Service) (ServiceMeta, error)
	Register(ctx context.Context, svc ServiceMeta) error
	DeRegister(ctx context.Context, svc ServiceMeta) error
}

type Event string

const (
	ADD Event = "add"
	DEL Event = "del"
	CHG Event = "chg"
)
