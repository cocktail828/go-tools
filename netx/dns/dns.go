package dns

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/cocktail828/go-tools/z/ttlmap"
)

type entry struct {
	err   error
	value any
	cname string
}

// DNS Resolver with cache
type Resolver struct {
	cache    *ttlmap.Cache[entry]
	Resolver net.Resolver
	TTL      time.Duration // DNS 记录缓存最长时间
	NegTTL   time.Duration // 如果非0, 则为 negative record 的缓存时间, 默认为0
}

func (r *Resolver) store(key string, err error, val any, cname string) {
	if err == nil {
		r.cache.SetWithTTL(key, entry{err, val, cname}, r.TTL)
	} else {
		r.cache.SetWithTTL(key, entry{err, val, cname}, r.NegTTL)
	}
}

func (r *Resolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	key := "LookupAddr#" + addr
	if val, err := r.cache.Get(key); err == nil {
		return val.value.([]string), val.err
	}

	res, err := r.Resolver.LookupAddr(ctx, addr)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	key := "LookupCNAME#" + host
	if val, err := r.cache.Get(key); err == nil {
		return val.value.(string), val.err
	}

	res, err := r.Resolver.LookupCNAME(ctx, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	key := "LookupHost#" + host
	if val, err := r.cache.Get(key); err == nil {
		return val.value.([]string), val.err
	}

	res, err := r.Resolver.LookupHost(ctx, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupIP(ctx context.Context, network string, host string) ([]net.IP, error) {
	key := "LookupIP#" + network + "#" + host
	if val, err := r.cache.Get(key); err == nil {
		return val.value.([]net.IP), val.err
	}

	res, err := r.Resolver.LookupIP(ctx, network, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	key := "LookupIPAddr#" + host
	if val, err := r.cache.Get(key); err == nil {
		return val.value.([]net.IPAddr), val.err
	}

	res, err := r.Resolver.LookupIPAddr(ctx, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	key := "LookupMX#" + name
	if val, err := r.cache.Get(key); err == nil {
		return val.value.([]*net.MX), val.err
	}

	res, err := r.Resolver.LookupMX(ctx, name)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	key := "LookupNS#" + name
	if val, err := r.cache.Get(key); err == nil {
		return val.value.([]*net.NS), val.err
	}

	res, err := r.Resolver.LookupNS(ctx, name)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupNetIP(ctx context.Context, network string, host string) ([]netip.Addr, error) {
	key := "LookupNetIP#" + network + "#" + host
	if val, err := r.cache.Get(key); err == nil {
		return val.value.([]netip.Addr), val.err
	}

	res, err := r.Resolver.LookupNetIP(ctx, network, host)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupPort(ctx context.Context, network string, service string) (port int, err error) {
	key := "LookupPort#" + network + "#" + service
	if val, err := r.cache.Get(key); err == nil {
		return val.value.(int), val.err
	}

	res, err := r.Resolver.LookupPort(ctx, network, service)
	r.store(key, err, res, "")
	return res, err
}

func (r *Resolver) LookupSRV(ctx context.Context, service string, proto string, name string) (string, []*net.SRV, error) {
	key := "LookupSRV#" + service + "#" + proto + "#" + name
	if val, err := r.cache.Get(key); err == nil {
		return val.cname, val.value.([]*net.SRV), val.err
	}

	cname, res, err := r.Resolver.LookupSRV(ctx, service, proto, name)
	r.store(key, err, res, cname)
	return cname, res, err
}

func (r *Resolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	key := "LookupTXT#" + name
	if val, err := r.cache.Get(key); err == nil {
		return val.value.([]string), val.err
	}

	res, err := r.Resolver.LookupTXT(ctx, name)
	r.store(key, err, res, "")
	return res, err
}
