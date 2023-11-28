package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNoSuchHost = errors.New("no such host")
)

type SRVSet struct {
	ips      []*net.SRV // SRV 解析结果
	createAt time.Time  // 记录创建的时间
	negative bool       // 空记录
}

func (r *SRVSet) expire(ttl, negttl time.Duration) bool {
	if r.negative {
		return time.Since(r.createAt) > negttl
	}
	return time.Since(r.createAt) > ttl
}

func (r *SRVSet) SetPort(port int) {
	for _, v := range r.ips {
		if v.Port == 0 {
			v.Port = uint16(port)
		}
	}
}

func (r *SRVSet) Equal(peer *SRVSet) bool {
	if len(r.ips) != len(peer.ips) {
		return false
	}
	oldMap := map[string]struct{}{}
	for _, v := range r.ips {
		oldMap[fmt.Sprintf("%v%v%v", v.Target, v.Port, v.Weight)] = struct{}{}
	}
	for _, v := range peer.ips {
		delete(oldMap, fmt.Sprintf("%v%v%v", v.Target, v.Port, v.Weight))
	}
	return len(oldMap) == 0
}

func (r *SRVSet) ToSrv() []*net.SRV {
	return r.ips
}

func (r *SRVSet) ToHostPort() []string {
	ips := make([]string, 0, len(r.ips))
	for _, ip := range r.ips {
		ips = append(ips, fmt.Sprintf("%v:%v", ip.Target, ip.Port))
	}
	return ips
}

type IPSet struct {
	ips      []net.IP  // 解析结果
	createAt time.Time // 记录使用的时间
	negative bool      // 空记录
}

func (r *IPSet) expire(ttl, negttl time.Duration) bool {
	if r.negative {
		return time.Since(r.createAt) > negttl
	}
	return time.Since(r.createAt) > ttl
}

func (r *IPSet) Equal(peer *IPSet) bool {
	if len(r.ips) != len(peer.ips) {
		return false
	}
	oldMap := map[string]struct{}{}
	for _, v := range r.ips {
		oldMap[v.String()] = struct{}{}
	}
	for _, v := range peer.ips {
		delete(oldMap, v.String())
	}
	return len(oldMap) == 0
}

func (r *IPSet) To4() []string {
	ips := make([]string, 0, len(r.ips))
	for _, ip := range r.ips {
		ips = append(ips, ip.To4().String())
	}
	return ips
}

func (r *IPSet) To16() []string {
	ips := make([]string, 0, len(r.ips))
	for _, ip := range r.ips {
		ips = append(ips, ip.To16().String())
	}
	return ips
}

func (r *IPSet) ToIP() []string {
	ips := make([]string, 0, len(r.ips))
	for _, ip := range r.ips {
		ips = append(ips, ip.String())
	}
	return ips
}

// DNS Resolver with cache
type Resolver struct {
	inited  atomic.Bool
	cache   sync.Map
	Timeout time.Duration // DNS 解析超时时间, 默认1s
	TTL     time.Duration // DNS 记录缓存最长时间
	NegTTL  time.Duration // 如果非0, 则为 negative record 的缓存时间, 默认为0
}

func (r *Resolver) normalize() {
	if !r.inited.CompareAndSwap(false, true) {
		return
	}
	if r.Timeout == 0 {
		r.Timeout = time.Second
	}
	if r.TTL == 0 {
		r.Timeout = time.Second * 15
	}
}

// Lookup IPv4 address
func (r *Resolver) LookupA(host string) (*IPSet, error) {
	r.normalize()
	rr, ok := r.cache.Load("ip4#" + host)
	if ok {
		if s := rr.(*IPSet); !s.expire(r.TTL, r.NegTTL) {
			return s, nil
		}
	}
	return r.lookupIP("ip4", host)
}

// Lookup IPv6 address
func (r *Resolver) LookupAAAA(host string) (*IPSet, error) {
	r.normalize()
	rr, ok := r.cache.Load("ip6#" + host)
	if ok {
		if s := rr.(*IPSet); !s.expire(r.TTL, r.NegTTL) {
			return s, nil
		}
	}
	return r.lookupIP("ip6", host)
}

// Lookup IPv4&IPv6 address
func (r *Resolver) LookupIP(host string) (*IPSet, error) {
	r.normalize()
	rr, ok := r.cache.Load("ip#" + host)
	if ok {
		if s := rr.(*IPSet); !s.expire(r.TTL, r.NegTTL) {
			return s, nil
		}
	}
	return r.lookupIP("ip", host)
}

func (r *Resolver) lookupIP(network, host string) (*IPSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	rr, err := net.DefaultResolver.LookupIP(ctx, network, host)
	if err != nil {
		if e, ok := err.(*net.DNSError); ok && r.NegTTL != 0 && e.Err == "no such host" {
			r.cache.Store(network+"#"+host, &IPSet{negative: true, createAt: time.Now()})
		}
		return nil, err
	}
	if len(rr) == 0 {
		return nil, ErrNoSuchHost
	}

	rlt := &IPSet{ips: rr, createAt: time.Now()}
	r.cache.Store(network+"#"+host, rlt)
	return rlt, nil
}

func (r *Resolver) LookupSRV(service, proto, name string) (*SRVSet, error) {
	r.normalize()
	rr, ok := r.cache.Load(service + "#" + proto + "#" + name)
	if ok {
		if s := rr.(*SRVSet); !s.expire(r.TTL, r.NegTTL) {
			return s, nil
		}
	}
	return r.lookupSRV(service, proto, name)
}

func (r *Resolver) lookupSRV(service, proto, name string) (*SRVSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	_, rr, err := net.DefaultResolver.LookupSRV(ctx, service, proto, name)
	if err != nil {
		if e, ok := err.(*net.DNSError); ok && r.NegTTL != 0 && e.Err == "no such host" {
			r.cache.Store(service+"#"+proto+"#"+name, &IPSet{negative: true, createAt: time.Now()})
		}
		return nil, err
	}
	if len(rr) == 0 {
		return nil, ErrNoSuchHost
	}

	rlt := &SRVSet{ips: rr, createAt: time.Now()}
	r.cache.Store(service+"#"+proto+"#"+name, rlt)
	return rlt, nil
}
