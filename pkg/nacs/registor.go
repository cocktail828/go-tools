package nacs

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"sync"

	"github.com/cocktail828/go-tools/pkg/nacs/registor"
)

var (
	ErrNoSuchRegistor = errors.New("no such registor")
)

var _ registor.Registor = &Registor{}

type Registor struct {
	cache   sync.Map
	regstor registor.Registor
	dumper  dumper
}

func NewRegistor(addr, group string) (*Registor, error) {
	r := &Registor{
		dumper: dumper{".findercache", "service_"},
	}

	urld, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	builder := registor.Lookup(urld.Scheme)
	if builder == nil {
		return nil, ErrNoSuchConfigor
	}
	r.regstor = builder.Build()
	return r, nil
}

type registorChanger struct {
	*Registor
	handler registor.Handler
}

func (c registorChanger) OnChange(evt registor.Event, meta registor.ServiceMeta, err error) {
	if err == nil {
		c.onEvent(evt, meta)
	}
	c.handler.OnChange(evt, meta, err)
}

func (r *Registor) Watch(ctx context.Context, handler registor.Handler, svcs ...registor.Service) error {
	return r.regstor.Watch(ctx, registorChanger{r, handler}, svcs...)
}

func (r *Registor) onEvent(evt registor.Event, meta registor.ServiceMeta) {
	switch evt {
	case registor.ADD, registor.CHG:
		r.cache.Store(meta.Service.String(), meta)
		data, _ := json.Marshal(meta)
		r.dumper.Dump(meta.Service.String(), data)

	case registor.DEL:
		r.cache.Delete(meta.Service.String())
		r.dumper.Remove(meta.Service.String())
	}
}

func (r *Registor) Lookup(svc registor.Service) (registor.ServiceMeta, error) {
	if meta, ok := r.cache.Load(svc.String()); ok {
		return meta.(registor.ServiceMeta), nil
	}

	meta, err := r.regstor.Lookup(svc)
	if err == nil {
		r.onEvent(registor.ADD, meta)
	}
	return meta, err
}

func (r *Registor) Register(ctx context.Context, svc registor.ServiceMeta) error {
	return r.regstor.Register(ctx, svc)
}

func (r *Registor) DeRegister(ctx context.Context, svc registor.ServiceMeta) error {
	return r.regstor.DeRegister(ctx, svc)
}
