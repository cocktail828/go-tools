package dns

import (
	"net"
	"strconv"
	"strings"
	"sync"
)

const (
	DNS_A_PREFIX   = "dns://"
	DNS_SRV_PREFIX = "dns+srv://"
)

type LBPolicy string

const (
	StaticRoundRobin        LBPolicy = "static_rr"   // 静态地址列表, 不检查地址有效性
	DynamicRoundRobin       LBPolicy = "dynamic_rr"  // DNS A 记录
	DynamicWeightRoundRobin LBPolicy = "dynamic_wrr" // DNS SRV 记录
)

// NOTICE:
// 域名可以拥有多个 A 记录
// 域名只允许设置一个 CNAME 记录, 但是每个记录的 target 必须是 A 地址
type RRSet struct {
	mu       sync.RWMutex
	name     string
	lbPolicy LBPolicy
	records  []net.SRV
}

func NewRRSet(name string) (*RRSet, error) {
	lhs := &RRSet{name: name}
	if err := lhs.Reset(name); err != nil {
		return nil, err
	}
	return lhs, nil
}

func (lhs *RRSet) Reset(name string) error {
	if name == "" {
		name = lhs.name
	}

	lbPolicy := StaticRoundRobin
	switch {
	case strings.HasPrefix(name, DNS_A_PREFIX):
		lbPolicy = DynamicRoundRobin
		name = name[len(DNS_A_PREFIX):]

	case strings.HasPrefix(name, DNS_SRV_PREFIX):
		lbPolicy = DynamicWeightRoundRobin
		name = name[len(DNS_SRV_PREFIX):]
	}

	records, err := lhs.lookup(name, lbPolicy)
	if err != nil {
		return err
	}

	lhs.mu.Lock()
	defer lhs.mu.Unlock()
	lhs.lbPolicy = lbPolicy
	lhs.name = name
	lhs.records = records
	return nil
}

func (lhs *RRSet) Normalize(port int) *RRSet {
	lhs.mu.Lock()
	defer lhs.mu.Unlock()
	for i := 0; i < len(lhs.records); i++ {
		if lhs.records[i].Port == 0 {
			lhs.records[i].Port = uint16(port)
		}
	}
	return lhs
}

func (lhs *RRSet) Name() string {
	lhs.mu.RLock()
	defer lhs.mu.RUnlock()
	return lhs.name
}

func (lhs *RRSet) LBPolicy() LBPolicy {
	lhs.mu.RLock()
	defer lhs.mu.RUnlock()
	return lhs.lbPolicy
}

func (lhs *RRSet) Records() []net.SRV {
	lhs.mu.RLock()
	defer lhs.mu.RUnlock()
	return lhs.records[:]
}

func (lhs *RRSet) lookup(name string, lbPolicy LBPolicy) ([]net.SRV, error) {
	records := []net.SRV{}

	switch lbPolicy {
	case DynamicRoundRobin:
		{
			host, port := splitHostPort(name)
			ips, err := LookupA(host)
			if err != nil {
				return nil, err
			}

			for i := 0; i < len(ips); i++ {
				records = append(records, net.SRV{
					Target: ips[i].String(),
					Port:   uint16(port),
				})
			}
		}

	case DynamicWeightRoundRobin:
		{
			_, addrs, err := net.LookupSRV("", "", name)
			for i := 0; i < len(addrs); i++ {
				if addrs[i] != nil {
					records = append(records, *(addrs[i]))
				}
			}
			return nil, err
		}

	default:
		{
			for _, subname := range strings.Split(name, ",") {
				host, port := splitHostPort(subname)
				records = append(records, net.SRV{
					Target: host,
					Port:   uint16(port),
				})
			}
		}
	}
	return records, nil
}

func splitHostPort(host string) (string, int) {
	h, p, err := net.SplitHostPort(host)
	if err != nil {
		return host, 0
	}
	pi, _ := strconv.Atoi(p)
	return h, pi
}
