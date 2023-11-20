package registry

import (
	"context"
	"sync"

	"github.com/cocktail828/go-tools/algo/balancer"
)

type svcEntries struct {
	entries []Entry
}

func (se *svcEntries) Len() int {
	return len(se.entries)
}

type Agent struct {
	register   Register
	configer   Configer
	deregister DeRegister
	lb         balancer.RoundRobin
	caller     func(context.Context, Entry, []byte) ([]byte, error)
	mu         sync.RWMutex
	cacheSvcs  map[string]*svcEntries
}

type Option func(*Agent)

func WithBalancer() Option {
	return func(a *Agent) {}
}

func WithRegister(reg Register) Option {
	return func(a *Agent) { a.register = reg }
}

func WithConfiger(cfg Configer) Option {
	return func(a *Agent) { a.configer = cfg }
}

func WithCaller(f func(context.Context, Entry, []byte) ([]byte, error)) Option {
	return func(a *Agent) { a.caller = f }
}

func New(opts ...Option) *Agent {
	a := &Agent{cacheSvcs: map[string]*svcEntries{}}
	for _, f := range opts {
		f(a)
	}
	ctx, cancel := context.WithCancel(context.Background())
	func() {
		go a.watchService(context.Background())
		cancel()
	}()
	<-ctx.Done()
	return a
}

func (a *Agent) Call(ctx context.Context, svc, ver string, body []byte) ([]byte, error) {
	a.mu.RLock()
	e, ok := a.cacheSvcs[svc+"#"+CheckVersion(ver)]
	a.mu.RUnlock()
	if !ok || e.Len() == 0 {
		return nil, ErrNoAvailNode
	}
	pos := a.lb.Get(e)
	if pos < 0 {
		return nil, ErrNoAvailNode
	}
	n := e.entries[pos]
	return a.caller(ctx, n, body)
}

func (a *Agent) watchService(ctx context.Context) {
	a.register.WatchServices(ctx, func(entries []Entry) {
		a.mu.Lock()
		defer a.mu.Unlock()
		for _, e := range entries {
			if v, ok := a.cacheSvcs[e.Name+"#"+e.Version]; ok {
				v.entries = append(v.entries, e)
			} else {
				a.cacheSvcs[e.Name+"#"+e.Version] = &svcEntries{entries: []Entry{e}}
			}
		}
	})
}

func (a *Agent) Register(ctx context.Context, reg Registration) error {
	if a.register == nil {
		return ErrNoRegister
	}

	dereger, err := a.register.Register(ctx, reg)
	if err == nil {
		a.deregister = dereger
	}
	return err
}

func (a *Agent) DeRegister(ctx context.Context) error {
	if a.deregister == nil {
		return ErrNotRegistered
	}
	return a.deregister.DeRegister(ctx)
}

func (a *Agent) Services(ctx context.Context, svc, ver string) ([]Entry, error) {
	if a.register == nil {
		return nil, ErrNoRegister
	}
	return a.register.Services(ctx, svc, ver)
}

func (a *Agent) WatchConfig(ctx context.Context, svc, ver string, f func(map[string][]byte)) error {
	if a.register == nil {
		return ErrNoConfiger
	}
	return a.configer.WatchConfig(ctx, svc, ver, f)
}

func (a *Agent) GetConfig(ctx context.Context, svc, ver string) (map[string][]byte, error) {
	if a.register == nil {
		return nil, ErrNoConfiger
	}
	return a.configer.Pull(ctx, svc, ver)
}
