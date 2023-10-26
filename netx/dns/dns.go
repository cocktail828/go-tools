package dns

import (
	"net"
	"strconv"
	"strings"
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
	name        string
	defaultPort int
	lbPolicy    LBPolicy
	records     []net.SRV
}

type Option func(*RRSet)

func WithPort(port int) Option {
	return func(r *RRSet) {
		r.defaultPort = port
	}
}

func NewRRSet(name string, opts ...Option) (*RRSet, error) {
	lhs := &RRSet{
		name:     name,
		lbPolicy: StaticRoundRobin,
	}
	for _, o := range opts {
		o(lhs)
	}

	switch {
	case strings.HasPrefix(name, DNS_A_PREFIX):
		lhs.lbPolicy = DynamicRoundRobin
		name = name[len(DNS_A_PREFIX):]

	case strings.HasPrefix(name, DNS_SRV_PREFIX):
		lhs.lbPolicy = DynamicWeightRoundRobin
		name = name[len(DNS_SRV_PREFIX):]
	}

	records, err := lhs.lookup(name, lhs.lbPolicy)
	if err != nil {
		return nil, err
	}

	lhs.records = records
	for i := 0; i < len(lhs.records); i++ {
		if lhs.records[i].Port == 0 {
			lhs.records[i].Port = uint16(lhs.defaultPort)
		}
	}

	return lhs, nil
}

func (lhs *RRSet) Name() string {
	return lhs.name
}

func (lhs *RRSet) LBPolicy() LBPolicy {
	return lhs.lbPolicy
}

func (lhs *RRSet) Records() []net.SRV {
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
