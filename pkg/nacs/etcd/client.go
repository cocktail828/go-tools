package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// etcd URI format: etcd://host:2379?namespace=$namespace&service=$service&version=$version
type EtcdClient struct {
	namespace  string
	service    string
	version    string
	ttl        int64
	baseClient *BaseEtcdClient
}

func (c *EtcdClient) Prefix() string {
	return fmt.Sprintf("%s/%s@%s", c.namespace, c.service, c.version)
}

// ServiceKey returns the service key in etcd format: $namespace/$service@$version/instances/$host:$port
// It is used to identify the service instance in etcd.
func (c *EtcdClient) ServiceKey(host string, port uint) string {
	return fmt.Sprintf("%s/instances/%s:%d", c.Prefix(), host, port)
}

// ConfigID returns the config ID in etcd format: $namespace/$service@$version/config
// It is used to identify the config in etcd.
func (c *EtcdClient) ConfigID() string {
	return fmt.Sprintf("%s/config", c.Prefix())
}

// NewEtcdClient creates a new etcd-backed client. The URL query should contain namespace, service, version
func NewEtcdClient(u *url.URL) (*EtcdClient, error) {
	baseClient, err := NewBaseEtcdClient(u)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	return baseClient.Spawn(query.Get("namespace"), query.Get("service"), query.Get("version"))
}

func (c *EtcdClient) Ancestor() *BaseEtcdClient {
	return c.baseClient
}

func (c *EtcdClient) Share(namespace, service, version string) (*EtcdClient, error) {
	return c.baseClient.Spawn(namespace, service, version)
}

func (c *EtcdClient) Close() error {
	return c.baseClient.Close()
}

func (c *EtcdClient) Register(host string, port uint, meta map[string]string) (context.CancelFunc, error) {
	inst := nacs.Instance{Name: c.service + "@" + c.version, Host: host, Port: port, Meta: meta}
	v, err := json.Marshal(inst)
	if err != nil {
		return nil, err
	}

	// create lease
	ctx := context.Background()
	lease, err := c.baseClient.client.Grant(ctx, c.ttl)
	if err != nil {
		return nil, err
	}

	if _, err := c.baseClient.client.Put(ctx, c.ServiceKey(host, port), string(v), clientv3.WithLease(lease.ID)); err != nil {
		return nil, err
	}

	// keep alive
	ch, err := c.baseClient.client.KeepAlive(ctx, lease.ID)
	if err != nil {
		return nil, err
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-cancelCtx.Done():
				c.baseClient.client.Revoke(context.Background(), lease.ID)
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
			}
		}
	}()

	return func() { cancel(); c.DeRegister(host, port) }, nil
}

func (c *EtcdClient) DeRegister(host string, port uint) error {
	_, err := c.baseClient.client.Delete(context.Background(), c.ServiceKey(host, port))
	return err
}

func (c *EtcdClient) Discover() ([]nacs.Instance, error) {
	resp, err := c.baseClient.client.Get(context.Background(), c.Prefix()+"/instances/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	res := make([]nacs.Instance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var inst nacs.Instance
		if err := json.Unmarshal(kv.Value, &inst); err != nil {
			continue
		}
		res = append(res, inst)
	}
	return res, nil
}

func (c *EtcdClient) Watch(callback func([]nacs.Instance, error)) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := c.baseClient.client.Watch(ctx, c.Prefix()+"/instances/", clientv3.WithPrefix(), clientv3.WithPrevKV())

	go func() {
		for wr := range ch {
			if wr.Err() != nil {
				callback(nil, wr.Err())
				continue
			}

			instances, _ := c.Discover()
			callback(instances, nil)
		}
	}()

	return cancel, nil
}

func (c *EtcdClient) Load() ([]byte, error) {
	resp, err := c.baseClient.client.Get(context.Background(), c.ConfigID())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, errors.Errorf("config %s not found", c.ConfigID())
	}
	return resp.Kvs[0].Value, nil
}

func (c *EtcdClient) Monitor(cb func(name string, payload []byte, err error)) (context.CancelFunc, error) {
	if cb == nil {
		return nil, errors.Errorf("callback is nil")
	}
	// watch config key
	ctx, cancel := context.WithCancel(context.Background())
	rch := c.baseClient.client.Watch(ctx, c.ConfigID())
	go func() {
		for wr := range rch {
			if wr.Err() != nil {
				cb(c.ConfigID(), nil, wr.Err())
				continue
			}
			for _, ev := range wr.Events {
				cb(c.ConfigID(), ev.Kv.Value, nil)
			}
		}
	}()

	return cancel, nil
}
