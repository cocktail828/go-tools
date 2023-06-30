package dns

import (
	"context"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/algo/loadbalance"
	"github.com/cocktail828/go-tools/z"
	"github.com/cocktail828/go-tools/z/diagnostic"
)

var (
	defaultLookupInterval = time.Second * 15
	defaultHealthInterval = time.Second * 3
	defaultProbeTimeout   = time.Second
)

const (
	DNS_A_PREFIX   = "dns://"
	DNS_SRV_PREFIX = "dns+srv://"
)

type LBPolicy string

const (
	StaticRoindRobin        LBPolicy = "static_rr"   // 静态地址列表, 不检查地址有效性
	DynamicRoindRobin       LBPolicy = "dynamic_rr"  // DNS A 记录
	DynamicWeightRoindRobin LBPolicy = "dynamic_wrr" // DNS net.SRV 记录
)

type Config struct {
	Port           int           // 必传, 用于端口探活
	LookupInterval time.Duration // DNS 地址查询间隔, 默认 15s
	ProbeInterval  time.Duration // 端口探活间隔, 默认 3s
	ProbeTimeout   time.Duration // 端口探活超时, 默认 1s
	Probe          Prober        // 探活接口
}

type RRManager struct {
	ctx       context.Context
	cancel    context.CancelFunc
	config    Config
	lb        *loadbalance.RoundRobin
	eventChan chan string
	mu        sync.RWMutex
	rrSet     *RRSet
}

func (lhs *RRManager) Len() int {
	if lhs.rrSet != nil {
		return len(lhs.rrSet.Records)
	}
	return 0
}

func (lhs *RRManager) Validate(idx int) bool {
	if lhs.rrSet == nil {
		return true
	}

	srv := &lhs.rrSet.Records[idx]
	now := time.Now()
	if srv.lastCheckAt.Add(time.Second).After(now) {
		return !srv.invalid
	}
	srv.lastCheckAt = now

	if err := lhs.config.Probe(
		net.JoinHostPort(srv.Target, strconv.Itoa(int(srv.Port))),
		lhs.config.ProbeTimeout,
	); err != nil {
		srv.invalid = true
	}
	return !srv.invalid
}

func (lhs *RRManager) Weight(idx int) int {
	if lhs.rrSet != nil {
		return int(lhs.rrSet.Records[idx].Weight)
	}
	return -1
}

func NewRRManager(cfg Config) *RRManager {
	if cfg.LookupInterval == 0 {
		cfg.LookupInterval = defaultLookupInterval
	}

	if cfg.ProbeInterval == 0 {
		cfg.ProbeInterval = defaultHealthInterval
	}

	if cfg.ProbeTimeout == 0 {
		cfg.ProbeTimeout = defaultProbeTimeout
	}

	if cfg.Probe == nil {
		cfg.Probe = func(addr string, tmo time.Duration) error { return nil }
	}

	ctx, cancel := context.WithCancel(context.Background())
	mgr := &RRManager{
		ctx:       ctx,
		cancel:    cancel,
		config:    cfg,
		lb:        loadbalance.NewRoundRobin(),
		eventChan: make(chan string, 100),
	}
	go mgr.discover()

	return mgr
}

// 刷新全局地址列表
func (lhs *RRManager) discover() {
	timer := time.NewTimer(lhs.config.LookupInterval)
	defer timer.Stop()

	for {
		select {
		case <-lhs.ctx.Done():
			return

		case <-timer.C:
			z.WithRLock(&lhs.mu, func() {
				if lhs.rrSet != nil && lhs.rrSet.Name != "" {
					lhs.eventChan <- lhs.rrSet.Name
				}
			})
			timer.Reset(lhs.config.LookupInterval)

		case name := <-lhs.eventChan:
			if rrset, err := NewRRSet(name); err == nil {
				z.WithLock(&lhs.mu, func() { lhs.rrSet = rrset.Normalize(lhs.config.Port) })
			}
		}
	}
}

func (lhs *RRManager) Get() *SRV {
	lhs.mu.RLock()
	defer lhs.mu.RUnlock()
	if lhs.rrSet != nil {
		if pos := lhs.lb.Get(lhs); pos != -1 {
			return &lhs.rrSet.Records[pos]
		}
	}
	return nil
}

func (lhs *RRManager) Endpoints() []SRV {
	lhs.mu.RLock()
	defer lhs.mu.RUnlock()
	if lhs.rrSet == nil {
		return nil
	}
	return lhs.rrSet.Records[:]
}

func (lhs *RRManager) Reset(name string) {
	if name != "" {
		lhs.eventChan <- name
	}
}

func lbPolicy(host string) LBPolicy {
	switch {
	case strings.HasPrefix(host, DNS_A_PREFIX):
		return DynamicRoindRobin

	case strings.HasPrefix(host, DNS_SRV_PREFIX):
		return DynamicWeightRoindRobin

	default:
		return StaticRoindRobin
	}
}

type SRV struct {
	net.SRV
	Name        string
	invalid     bool
	lastCheckAt time.Time
}

// NOTICE:
// 域名可以拥有多个 A 记录
// 域名只允许设置一个 CNAME 记录
// 域名可以拥有多条 net.SRV 记录, 但是每个记录的 target 必须是 A 地址
type RRSet struct {
	Name     string   `yaml:"name,omitempty"`
	LBPolicy LBPolicy `yaml:"lb_policy,omitempty"`
	Records  []SRV    `yaml:"records,omitempty"`
}

func (lhs *RRSet) Normalize(port int) *RRSet {
	for i := 0; i < len(lhs.Records); i++ {
		if lhs.Records[i].Port == 0 {
			lhs.Records[i].Port = uint16(port)
		}
	}
	return lhs
}

func NewRRSet(name string) (*RRSet, error) {
	lhs := &RRSet{
		Name:     name,
		LBPolicy: lbPolicy(name),
	}

	switch lhs.LBPolicy {
	case DynamicRoindRobin:
		name = lhs.Name[len(DNS_A_PREFIX):]
	case DynamicWeightRoindRobin:
		name = lhs.Name[len(DNS_SRV_PREFIX):]
	}

	diag := diagnostic.New()
	for _, name := range strings.Split(name, ",") {
		if err := lhs.lookup(name); err != nil {
			diag = diag.WithError(err)
		}
	}

	if diag.HasError() {
		return nil, diag.ToError()
	}
	return lhs, nil
}

func (lhs *RRSet) lookup(name string) error {
	switch lhs.LBPolicy {
	case DynamicRoindRobin:
		host, port := splitHostPort(name)
		ips, err := LookupA(host)
		if err != nil {
			return err
		}

		for i := 0; i < len(ips); i++ {
			lhs.Records = append(lhs.Records, SRV{
				SRV: net.SRV{
					Target: ips[i].String(),
					Port:   uint16(port),
				},
				Name: name,
			})
		}

	case DynamicWeightRoindRobin:
		_, addrs, err := net.LookupSRV("", "", name)
		for i := 0; i < len(addrs); i++ {
			if addrs[i] != nil {
				lhs.Records = append(lhs.Records, SRV{
					SRV:  *(addrs[i]),
					Name: name,
				})
			}
		}
		return err

	default:
		host, port := splitHostPort(name)
		lhs.Records = append(lhs.Records, SRV{
			SRV: net.SRV{
				Target: host,
				Port:   uint16(port),
			},
			Name: name,
		})
	}
	return nil
}

func splitHostPort(host string) (string, int) {
	h, p, err := net.SplitHostPort(host)
	if err != nil {
		return host, 0
	}
	pi, _ := strconv.Atoi(p)
	return h, pi
}
