package registry

import (
	"context"
	"errors"
	"log"

	"github.com/cocktail828/go-tools/netx/inet"
	"github.com/cocktail828/go-tools/z"
)

type Registration struct {
	Name    string
	Version string
	Address string
	Port    int
	Meta    map[string]string
}

func (r *Registration) Normalize() error {
	if r.Version == "" {
		return ErrMissingVersion
	}
	if r.Port == 0 {
		return ErrMissingPort
	}
	if r.Name == "" {
		return ErrMissingName
	}
	if r.Address == "" {
		addrs, err := inet.Inet4()
		z.Must(err)
		if len(addrs) == 0 {
			z.Must(errors.New("no public v4 addr found"))
		}
		r.Address = addrs[0].String()
		log.Println("[WARN] using default address", r.Address)
	}
	return nil
}

type Entry struct {
	Name    string
	Version string
	Address string
	Port    int
	Meta    map[string]string
}

type DeRegister interface {
	DeRegister(ctx context.Context) error
}

type Register interface {
	Register(ctx context.Context, sc Registration) (DeRegister, error)
	Services(ctx context.Context, svc, ver string) ([]Entry, error)
	WatchService(ctx context.Context, svc, ver string, cb func(entries []Entry)) error
	WatchServices(ctx context.Context, cb func(entries []Entry)) error
}

type Configer interface {
	Pull(ctx context.Context, svc, ver string) (map[string][]byte, error)
	WatchConfig(ctx context.Context, svc, ver string, cb func(map[string][]byte)) error
}

type Event struct {
	Kind int
	Meta map[string]string
	Body []byte
}

type EventEngine interface {
	Fire(ctx context.Context, svc, ver, name string, e Event) error
	Recv(ctx context.Context, name string) ([]Event, error)
}
