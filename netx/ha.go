package netx

import (
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/cocktail828/go-tools/werror"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	DNS_A_PREFIX   = "dns://"
	DNS_SRV_PREFIX = "dns+srv://"
)

type LBPolicy string

const (
	StaticRoindRobin        LBPolicy = "static_rr"   // 静态地址列表, 不检查地址有效性
	DynamicRoindRobin       LBPolicy = "dynamic_rr"  // DNS A 记录
	DynamicWeightRoindRobin LBPolicy = "dynamic_wrr" // DNS SRV 记录
)

func dynamicA(addr string) bool {
	return strings.HasPrefix(addr, DNS_A_PREFIX)
}

func dynamicSRV(addr string) bool {
	return strings.HasPrefix(addr, DNS_SRV_PREFIX)
}

type RRSet struct {
	Logger    *logrus.Logger
	CacheFile string // 本地记录缓存文件
	addr      string
	mu        sync.RWMutex
	lbPolicy  LBPolicy          // 负载策略
	rrSets    map[string]*rrSet // DNS 解析后的地址
}

func NewRRSet(addr string, cf string) *RRSet {
	if cf == "" {
		cf = "dns.cache"
	}

	rrset := &RRSet{
		Logger:    logrus.New(),
		CacheFile: cf,
		addr:      addr,
		lbPolicy:  lbPolicyF(addr),
		rrSets:    make(map[string]*rrSet),
	}
	rrset.Logger.SetLevel(logrus.ErrorLevel)
	rrset.buildLocked(addr)

	return rrset
}

func (rrset *RRSet) LBPolicy() LBPolicy {
	return rrset.lbPolicy
}

func (rrset *RRSet) Endpoints() []RR {
	rrset.mu.RLock()
	defer rrset.mu.RUnlock()

	rlt := make([]RR, 0, len(rrset.rrSets))
	for _, rrs := range rrset.rrSets {
		for _, rr := range rrs.RRs {
			rlt = append(rlt, rr.clone())
		}
	}
	return rlt
}

func (rrset *RRSet) addLocked(name string) error {
	rrs, err := rrSetFromName(rrset.lbPolicy, escape(name))
	if err != nil {
		return err
	}
	rrset.rrSets[name] = rrs
	return nil
}

// changed?
// generally addr should be "", or will use `addr`.
// if addr is invalid, will use the old addr
func (rrset *RRSet) Refresh(addr string) bool {
	rrset.mu.Lock()
	defer rrset.mu.Unlock()

	if addr != rrset.addr {
		changed, _ := rrset.buildLocked(addr)
		return changed
	}

	changed := false
	for _, rrs := range rrset.rrSets {
		if _changed, _ := rrs.refresh(); _changed {
			changed = true
		}
	}

	return changed
}

// changed?
func (rrset *RRSet) buildLocked(addr string) (bool, error) {
	oldset := rrset.rrSets
	oldlbPolicy := rrset.lbPolicy

	rrset.rrSets = make(map[string]*rrSet)
	rrset.lbPolicy = lbPolicyF(addr)
	for _, a := range splitAddr(addr) {
		rrset.addLocked(a)
	}

	if len(rrset.rrSets) != 0 {
		rrset.Logger.Debugf("rrset refreshed...")
		rrset.addr = addr
		rrset.dump(rrset.CacheFile)
		return true, nil
	}

	if len(oldset) != 0 {
		rrset.Logger.Debugf("rrset failback...")
		rrset.lbPolicy = oldlbPolicy
		rrset.rrSets = oldset
		return false, nil
	}

	names := []string{}
	err := rrset.load(rrset.CacheFile)
	if err == nil {
		for n := range rrset.rrSets {
			names = append(names, n)
		}
		rrset.addr = strings.Join(names, ",")
	}
	rrset.Logger.Debugf("rrset load from cache...")
	return true, err
}

func (rrset *RRSet) dump(fname string) error {
	body, err := yaml.Marshal(rrset.rrSets)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fname, body, os.ModePerm|fs.FileMode(os.O_TRUNC))
}

func (rrset *RRSet) load(fname string) error {
	body, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(body, &rrset.rrSets)
}

type rrSet struct {
	LBPolicy    LBPolicy       `yaml:"lbpolicy,omitempty"`
	EscapedName string         `yaml:"escaped,omitempty"` // 域名
	RRs         map[string]*RR `yaml:"rrs,omitempty"`     // 资源记录, map[target]...
}

func rrSetFromName(lbPolicy LBPolicy, escaped string) (*rrSet, error) {
	switch lbPolicy {
	case DynamicRoindRobin:
		target, port := splitHostPort(escaped)
		r, err := rrFromA(target)
		if err != nil {
			return nil, err
		}
		r.Port = port
		return &rrSet{
			EscapedName: target,
			LBPolicy:    lbPolicy,
			RRs:         map[string]*RR{target: r},
		}, nil

	case DynamicWeightRoindRobin:
		r, err := rrFromSRV("", "", escaped)
		if err != nil {
			return nil, err
		}
		return &rrSet{
			EscapedName: escaped,
			LBPolicy:    lbPolicy,
			RRs:         r,
		}, nil

	default:
		target, port := splitHostPort(escaped)
		return &rrSet{
			EscapedName: target,
			LBPolicy:    lbPolicy,
			RRs: map[string]*RR{
				target: {
					Hosts: []string{target},
					Port:  port,
				}},
		}, nil
	}
}

// changed?
func (rrs *rrSet) refresh() (bool, error) {
	nrrs, err := rrSetFromName(rrs.LBPolicy, rrs.EscapedName)
	if err != nil {
		return false, err
	}

	m := map[string]struct{}{}
	for target := range rrs.RRs {
		m[target] = struct{}{}
	}

	for target, nr := range nrrs.RRs {
		if or, ok := rrs.RRs[target]; ok && !or.compare(nr) {
			delete(m, target)
		}
	}

	if len(m) == 0 {
		return false, nil
	}

	rrs.RRs = nrrs.RRs
	return true, nil
}

// Resource record set，RRSet
type RR struct {
	Hosts  []string `yaml:"hosts,omitempty"`  // 多个 A 地址
	Port   string   `yaml:"port,omitempty"`   // 端口号. SRV 记录则使用域名中的值, 否则需要外部设置默认端口
	Weight int      `yaml:"weight,omitempty"` // 权重
}

func rrFromA(name string) (*RR, error) {
	hosts, err := net.LookupHost(name)
	if err != nil {
		return nil, err
	}
	return &RR{Hosts: hosts}, nil
}

// 查询 SRV 记录及所有 target, map[target]...
func rrFromSRV(service, proto, name string) (map[string]*RR, error) {
	_, srvs, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return nil, err
	}

	werr := werror.WrapperError{}
	rrset := map[string]*RR{}
	for _, srv := range srvs {
		r, err := rrFromA(srv.Target)
		if err != nil {
			werr.Add(err)
			continue
		}

		r.Port = strconv.Itoa(int(srv.Port))
		r.Weight = int(srv.Weight)
		rrset[srv.Target] = r
	}

	if len(rrset) != 0 {
		return rrset, nil
	}
	return nil, werr.Error()
}

// changed?
func (r *RR) compare(nr *RR) bool {
	m := make(map[string]struct{}, len(r.Hosts))
	for _, host := range r.Hosts {
		m[host] = struct{}{}
	}

	for _, host := range nr.Hosts {
		delete(m, host)
	}

	return len(m) != 0
}

func (r RR) clone() RR {
	nr := RR{
		Hosts:  make([]string, len(r.Hosts)),
		Port:   r.Port,
		Weight: r.Weight,
	}
	copy(nr.Hosts, r.Hosts)
	return nr
}

func splitAddr(addr string) []string {
	addrs := strings.Split(addr, ",")
	if len(addrs) <= 1 {
		return addrs
	}

	for _, a := range addrs {
		if dynamicA(a) || dynamicSRV(a) {
			panic("cannot use DNS loadbalance and host:port simultaneously")
		}
	}
	return addrs
}

func splitHostPort(addr string) (string, string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, ""
	}
	return host, port
}

func escape(name string) string {
	switch {
	case dynamicA(name):
		return strings.Trim(name, DNS_A_PREFIX)

	case dynamicSRV(name):
		return strings.Trim(name, DNS_SRV_PREFIX)
	default:
		return name
	}
}

func lbPolicyF(name string) LBPolicy {
	if dynamicA(name) {
		return DynamicRoindRobin
	}

	if dynamicSRV(name) {
		return DynamicWeightRoindRobin
	}
	return StaticRoindRobin
}
