package etcd

import (
	"context"
	"strings"
	"time"

	"github.com/cocktail828/go-tools/pkg/kvstore"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	DefaultPrefix = "/caesar/kvstore/"
)

// Currently a single etcd reader.
type etcd struct {
	prefix string
	client *clientv3.Client
}

func NewSource(opts ...kvstore.Option) (kvstore.KV, error) {
	options := kvstore.NewOptions(opts...)
	var endpoints []string
	if addrs, ok := options.Context.Value(addressKey{}).([]string); ok {
		endpoints = addrs
	}

	// check dial timeout option
	dialTimeout, ok := options.Context.Value(dialTimeoutKey{}).(time.Duration)
	if !ok {
		dialTimeout = 3 * time.Second // default dial timeout
	}

	config := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}
	if u, ok := options.Context.Value(authKey{}).(*authCreds); ok {
		config.Username = u.Username
		config.Password = u.Password
	}

	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	prefix := DefaultPrefix
	f, ok := options.Context.Value(prefixKey{}).(string)
	if ok {
		prefix = f
	}

	return &etcd{
		prefix: prefix,
		client: client,
	}, nil
}

func (c *etcd) Close() error {
	return c.client.Close()
}

func (c *etcd) String() string {
	return "etcd"
}

func (c *etcd) Delete(ctx context.Context, key string, opts ...kvstore.Option) error {
	options := kvstore.NewOptions(opts...)
	ecopts := []clientv3.OpOption{}

	if val := options.Context.Value(matchPrefix{}); val != nil {
		ecopts = append(ecopts, clientv3.WithPrefix())
	}
	_, err := c.client.Delete(ctx, c.prefix+key, ecopts...)
	return err
}

func (c *etcd) Write(ctx context.Context, key string, data []byte, opts ...kvstore.Option) error {
	_, err := c.client.Put(ctx, c.prefix+key, string(data))
	return err
}

func (c *etcd) Read(ctx context.Context, key string, opts ...kvstore.Option) ([]kvstore.KVPair, error) {
	options := kvstore.NewOptions(opts...)
	ecopts := []clientv3.OpOption{}

	if val := options.Context.Value(matchPrefix{}); val != nil {
		ecopts = append(ecopts, clientv3.WithPrefix())
	}

	rsp, err := c.client.Get(ctx, c.prefix+key, ecopts...)
	if err != nil {
		return nil, err
	}

	if rsp == nil {
		return nil, errors.Errorf("kvstore not found: %s/%s", c.prefix, key)
	}

	kvs := []kvstore.KVPair{}
	for _, kv := range rsp.Kvs {
		kvs = append(kvs, kvstore.KVPair{
			Key: strings.TrimPrefix(string(kv.Key), c.prefix),
			Val: kv.Value,
		})
	}
	return kvs, nil
}

func (c *etcd) Watch(ctx context.Context, key string, opts ...kvstore.Option) kvstore.Watcher {
	options := kvstore.NewOptions(opts...)
	ecopts := []clientv3.OpOption{clientv3.WithCreatedNotify()}

	if val := options.Context.Value(matchPrefix{}); val != nil {
		ecopts = append(ecopts, clientv3.WithPrefix())
	}

	sctx, cancel := context.WithCancel(ctx)
	return &watcher{
		ctx:        sctx,
		cancel:     cancel,
		prefix:     c.prefix,
		notifyChan: c.client.Watch(ctx, c.prefix+key, ecopts...),
	}
}
