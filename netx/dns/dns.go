package dns

import (
	"context"
	"net"
	"net/netip"
	"sync"
	"time"
)

// DNS Resolver with cache
type Resolver struct {
	cache    sync.Map
	Resolver net.Resolver
	TTL      time.Duration // DNS 记录缓存最长时间
	NegTTL   time.Duration // 如果非0, 则为 negative record 的缓存时间, 默认为0
}

type record struct {
	expireAt time.Time
	err      error
	value    any
	cname    string
}

func (r record) valid() bool { return r.expireAt.Before(time.Now()) }
func (r *Resolver) store(key string, err error, val any, cname string) {
	if err == nil && r.TTL != 0 {
		r.cache.Store(key, record{time.Now().Add(r.TTL), err, val, cname})
	}

	if err != nil && r.NegTTL != 0 {
		r.cache.Store(key, record{time.Now().Add(r.NegTTL), err, val, cname})
	}
}

func (r *Resolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	key := "LookupAddr#" + addr
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.([]string), t.err
			}
		}
	}

	res, err := r.Resolver.LookupAddr(ctx, addr)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	key := "LookupCNAME#" + host
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.(string), t.err
			}
		}
	}

	res, err := r.Resolver.LookupCNAME(ctx, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	key := "LookupHost#" + host
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.([]string), t.err
			}
		}
	}

	res, err := r.Resolver.LookupHost(ctx, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupIP(ctx context.Context, network string, host string) ([]net.IP, error) {
	key := "LookupIP#" + network + "#" + host
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.([]net.IP), t.err
			}
		}
	}

	res, err := r.Resolver.LookupIP(ctx, network, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	key := "LookupIPAddr#" + host
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.([]net.IPAddr), t.err
			}
		}
	}

	res, err := r.Resolver.LookupIPAddr(ctx, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	key := "LookupMX#" + name
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.([]*net.MX), t.err
			}
		}
	}

	res, err := r.Resolver.LookupMX(ctx, name)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	key := "LookupNS#" + name
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.([]*net.NS), t.err
			}
		}
	}

	res, err := r.Resolver.LookupNS(ctx, name)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupNetIP(ctx context.Context, network string, host string) ([]netip.Addr, error) {
	key := "LookupNetIP#" + network + "#" + host
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.([]netip.Addr), t.err
			}
		}
	}

	res, err := r.Resolver.LookupNetIP(ctx, network, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupPort(ctx context.Context, network string, service string) (port int, err error) {
	key := "LookupPort#" + network + "#" + service
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.(int), t.err
			}
		}
	}

	res, err := r.Resolver.LookupPort(ctx, network, service)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupSRV(ctx context.Context, service string, proto string, name string) (string, []*net.SRV, error) {
	key := "LookupSRV#" + service + "#" + proto + "#" + name
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.cname, t.value.([]*net.SRV), t.err
			}
		}
	}

	cname, res, err := r.Resolver.LookupSRV(ctx, service, proto, name)
	r.store(key, err, res, cname)
	return cname, res, err
}

func (r *Resolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	key := "LookupTXT#" + name
	if r.TTL != 0 {
		if val, ok := r.cache.Load(key); ok {
			if t := val.(record); t.valid() {
				return t.value.([]string), t.err
			}
		}
	}

	res, err := r.Resolver.LookupTXT(ctx, name)
	r.store(key, err, res, "")
	return res, err
}
