package nacs

import "context"

type Service struct {
	Name    string
	Version string
	Addr    string
	Meta    map[string]string
}

// TODO: implement
type ServiceHandler interface {
	OnServiceChange(ev Event, svc Service)
}

type DeRegister func(context.Context) error

// true: for pass
type Filter func(svc Service) bool
type Registry interface {
	Register(ctx context.Context, addr, ver string) (DeRegister, error)
	LookupService(ctx context.Context, options ...LookupAndWatchServiceOption) ([]Service, error)
	WatchService(ctx context.Context, handler ServiceHandler, options ...LookupAndWatchServiceOption) error
}

type LookupAndWatchServiceOption func(context.Context) context.Context
