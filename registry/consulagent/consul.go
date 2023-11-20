package consulagent

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/cocktail828/go-tools/registry"
	"github.com/cocktail828/go-tools/z/stringx"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

var _ registry.Register = &ConsulAgent{}
var _ registry.Configer = &ConsulAgent{}
var _ registry.EventEngine = &ConsulAgent{}

type ConsulAgent struct {
	client     *consulapi.Client
	svcWatcher *watch.Plan
	keyWatcher *watch.Plan
	svcID      string
	consulAddr string
}

func New(consulAddr string) *ConsulAgent {
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}
	return &ConsulAgent{client: client, consulAddr: consulAddr}
}

func (ca *ConsulAgent) Register(_ context.Context, r registry.Registration) (registry.DeRegister, error) {
	if err := r.Normalize(); err != nil {
		return nil, err
	}
	r.Version = registry.CheckVersion(r.Version)
	if r.Meta == nil {
		r.Meta = map[string]string{}
	}
	r.Meta["version"] = r.Version
	registration := &consulapi.AgentServiceRegistration{
		ID:      net.JoinHostPort(r.Address, strconv.Itoa(r.Port)) + "#" + stringx.RandomName(), // 服务节点的名称
		Name:    combine(r.Name, r.Version),                                                     // 服务名称
		Tags:    []string{r.Version},                                                            // tags
		Port:    r.Port,                                                                         // 服务端口
		Address: r.Address,                                                                      // 服务 IP 要确保consul可以访问这个ip
		Meta:    r.Meta,
	}

	registration.Check = &consulapi.AgentServiceCheck{ // 增加consul健康检查回调函数
		TCP:                            net.JoinHostPort(r.Address, strconv.Itoa(r.Port)),
		Timeout:                        "1s",
		Interval:                       "5s",  // 健康检查间隔
		DeregisterCriticalServiceAfter: "30s", // 故障检查失败30s后 consul自动将注册服务删除
	}

	if err := ca.client.Agent().ServiceRegister(registration); err != nil {
		return nil, err
	}
	ca.svcID = registration.ID
	return ca, nil
}

func (ca *ConsulAgent) DeRegister(_ context.Context) error {
	return ca.client.Agent().ServiceDeregister(ca.svcID)
}

func (ca *ConsulAgent) DeRegisterWith(_ context.Context, svcid string) error {
	return ca.client.Agent().ServiceDeregister(svcid)
}

func (ca *ConsulAgent) Services(_ context.Context, svc, ver string) ([]registry.Entry, error) {
	if ver == "" {
		return nil, registry.ErrMissingVersion
	}
	ver = registry.CheckVersion(ver)
	services, _, err := ca.client.Health().Service(combine(svc, ver), ver, false,
		&consulapi.QueryOptions{
			AllowStale:   true,
			MaxAge:       time.Second * 3,
			StaleIfError: time.Second * 15,
		})
	if err != nil {
		return nil, err
	}

	svcs := []registry.Entry{}
	for _, entry := range services {
		svcs = append(svcs, registry.Entry{
			Name:    svc,
			Address: entry.Service.Address,
			Port:    entry.Service.Port,
			Version: entry.Service.Meta["version"],
			Meta:    entry.Service.Meta,
		})
	}
	return svcs, nil
}

func (ca *ConsulAgent) WatchService(ctx context.Context, svc, ver string, cb func(entries []registry.Entry)) error {
	wp, err := watch.Parse(map[string]interface{}{
		"type":    "service",
		"service": combine(svc, registry.CheckVersion(ver)),
	})
	if err != nil {
		return err
	}
	ca.svcWatcher = wp
	wp.Handler = func(idx uint64, data interface{}) {
		switch d := data.(type) {
		case []*consulapi.ServiceEntry:
			entries := []registry.Entry{}
			for _, e := range d {
				entries = append(entries, registry.Entry{
					Name:    e.Service.Service,
					Address: e.Service.Address,
					Port:    e.Service.Port,
					Meta:    e.Service.Meta,
				})
			}
			cb(entries)
		}
	}
	return ca.svcWatcher.Run(ca.consulAddr)
}

func (ca *ConsulAgent) WatchServices(ctx context.Context, cb func(entries []registry.Entry)) error {
	wp, err := watch.Parse(map[string]interface{}{
		"type": "services",
	})
	if err != nil {
		return err
	}
	ca.svcWatcher = wp
	wp.Handler = func(idx uint64, data interface{}) {
		switch d := data.(type) {
		case map[string][]string:
			entries := []registry.Entry{}
			for svc := range d {
				arr := strings.Split(svc, "#")
				if len(arr) != 2 {
					continue
				}
				if tmps, err := ca.Services(ctx, arr[0], arr[1]); err == nil {
					entries = append(entries, tmps...)
				}
			}
			cb(entries)
		}
	}
	return ca.svcWatcher.Run(ca.consulAddr)
}

func (ca *ConsulAgent) StopWatchService() {
	if w := ca.svcWatcher; w != nil {
		w.Stop()
	}
}

func (ca *ConsulAgent) Pull(ctx context.Context, svc, ver string) (map[string][]byte, error) {
	prefix := svc + "/" + registry.CheckVersion(ver)
	kvpairs, _, err := ca.client.KV().List(prefix, nil)
	if err != nil {
		return nil, err
	}

	kvs := make(map[string][]byte, len(kvpairs))
	for _, pair := range kvpairs {
		kvs[strings.TrimLeft(pair.Key, prefix)] = pair.Value
	}
	return kvs, nil
}

func (ca *ConsulAgent) WatchConfig(ctx context.Context, svc, ver string, cb func(map[string][]byte)) error {
	wp, err := watch.Parse(map[string]interface{}{
		"type":   "keyprefix",
		"prefix": svc + "/" + registry.CheckVersion(ver),
	})
	if err != nil {
		return err
	}
	ca.keyWatcher = wp
	wp.Handler = func(idx uint64, data interface{}) {
		switch d := data.(type) {
		case consulapi.KVPairs:
			kvs := make(map[string][]byte, len(d))
			for _, pair := range d {
				kvs[pair.Key] = pair.Value
			}
			cb(kvs)
		}
	}
	return ca.keyWatcher.Run(ca.consulAddr)
}

func (ca *ConsulAgent) StopWatchConfig() {
	if w := ca.keyWatcher; w != nil {
		w.Stop()
	}
}

func (ca *ConsulAgent) Fire(_ context.Context, svc, ver, name string, e registry.Event) error {
	bs, _ := json.Marshal(e)
	_, _, err := ca.client.Event().Fire(&consulapi.UserEvent{
		Name:          name,
		Payload:       bs,
		ServiceFilter: combine(svc, registry.CheckVersion(ver)),
		TagFilter:     registry.CheckVersion(ver),
	}, nil)
	return err
}

func (ca *ConsulAgent) Recv(_ context.Context, name string) ([]registry.Event, error) {
	es, _, err := ca.client.Event().List(name, nil)
	if err != nil {
		return nil, err
	}
	aes := make([]registry.Event, 0, len(es))
	for _, e := range es {
		v := registry.Event{}
		if err := json.Unmarshal(e.Payload, &v); err == nil {
			aes = append(aes, v)
		}
	}
	return aes, nil
}
