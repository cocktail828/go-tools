package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNoSuchHost = errors.New("no such host")
)

type SRVSet struct {
	ips      []*net.SRV // SRV 解析结果
	accessAt time.Time  // 记录使用的时间
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
	r.accessAt = time.Now()
	return r.ips
}

func (r *SRVSet) ToHostPort() []string {
	r.accessAt = time.Now()
	ips := make([]string, 0, len(r.ips))
	for _, ip := range r.ips {
		ips = append(ips, fmt.Sprintf("%v:%v", ip.Target, ip.Port))
	}
	return ips
}

type IPSet struct {
	ips      []net.IP  // 解析结果
	accessAt time.Time // 记录使用的时间
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
	r.accessAt = time.Now()
	ips := make([]string, 0, len(r.ips))
	for _, ip := range r.ips {
		ips = append(ips, ip.To4().String())
	}
	return ips
}

func (r *IPSet) To16() []string {
	r.accessAt = time.Now()
	ips := make([]string, 0, len(r.ips))
	for _, ip := range r.ips {
		ips = append(ips, ip.To16().String())
	}
	return ips
}

func (r *IPSet) ToIP() []string {
	r.accessAt = time.Now()
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
	Timeout time.Duration // DNS 解析超时时间
	Refresh time.Duration // DNS 解析刷新间隔
	TTL     time.Duration // DNS 记录缓存最长时间
}

func (r *Resolver) startRefresher(refresh time.Duration) {
	if !r.inited.CompareAndSwap(false, true) {
		return
	}

	if r.Timeout == 0 {
		r.Timeout = time.Second * 3
	}
	if r.Refresh == 0 {
		r.Refresh = time.Second * 15
	}
	if r.TTL == 0 {
		r.Timeout = time.Hour
	}

	go func() {
		for {
			now := time.Now()
			addrMap := make(map[string]time.Time)
			r.cache.Range(func(key, value any) bool {
				if v, ok := value.(*IPSet); ok {
					addrMap[key.(string)] = v.accessAt
				}
				if v, ok := value.(*SRVSet); ok {
					addrMap[key.(string)] = v.accessAt
				}
				return true
			})

			for key, accessAt := range addrMap {
				if !accessAt.IsZero() && accessAt.Add(r.TTL).After(now) {
					r.cache.Delete(key)
				}

				arr := strings.Split(key, "#")
				switch len(arr) {
				case 2:
					r.lookupIP(arr[0], arr[1])
				case 3:
					r.lookupSRV(arr[0], arr[1], arr[2])
				default:
					r.cache.Delete(key)
				}
			}
			time.Sleep(r.Refresh)
		}
	}()
}

// Lookup IPv4 address
func (r *Resolver) LookupA(host string) (*IPSet, error) {
	r.startRefresher(r.Refresh)
	rr, ok := r.cache.Load("ip4#" + host)
	if ok {
		return rr.(*IPSet), nil
	}
	return r.lookupIP("ip4", host)
}

// Lookup IPv6 address
func (r *Resolver) LookupAAAA(host string) (*IPSet, error) {
	r.startRefresher(r.Refresh)
	rr, ok := r.cache.Load("ip6#" + host)
	if ok {
		return rr.(*IPSet), nil
	}
	return r.lookupIP("ip6", host)
}

// Lookup IPv4&IPv6 address
func (r *Resolver) LookupIP(host string) (*IPSet, error) {
	r.startRefresher(r.Refresh)
	rr, ok := r.cache.Load("ip#" + host)
	if ok {
		return rr.(*IPSet), nil
	}
	return r.lookupIP("ip", host)
}

func (r *Resolver) lookupIP(network, host string) (*IPSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	rr, err := net.DefaultResolver.LookupIP(ctx, network, host)
	if err != nil {
		return nil, err
	}
	if len(rr) == 0 {
		return nil, ErrNoSuchHost
	}

	rlt := &IPSet{ips: rr}
	r.cache.Store(network+"#"+host, rlt)
	return rlt, nil
}

func (r *Resolver) LookupSRV(service, proto, name string) (*SRVSet, error) {
	r.startRefresher(r.Refresh)
	rr, ok := r.cache.Load(service + "#" + proto + "#" + name)
	if ok {
		return rr.(*SRVSet), nil
	}
	return r.lookupSRV(service, proto, name)
}

func (r *Resolver) lookupSRV(service, proto, name string) (*SRVSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	_, rr, err := net.DefaultResolver.LookupSRV(ctx, service, proto, name)
	if err != nil {
		return nil, err
	}
	if len(rr) == 0 {
		return nil, ErrNoSuchHost
	}

	rlt := &SRVSet{ips: rr}
	r.cache.Store(service+"#"+proto+"#"+name, rlt)
	return rlt, nil
}
