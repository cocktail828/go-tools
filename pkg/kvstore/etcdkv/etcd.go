package etcdkv

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

func New(opts ...kvstore.Option) (kvstore.KV, error) {
	ctx := context.TODO()
	for _, o := range opts {
		ctx = o(ctx)
	}

	var endpoints []string
	if addrs, ok := ctx.Value(addressKey{}).([]string); ok {
		endpoints = addrs
	}

	// check dial timeout option
	dialTimeout, ok := ctx.Value(dialTimeoutKey{}).(time.Duration)
	if !ok {
		dialTimeout = 3 * time.Second // default dial timeout
	}

	config := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}
	if u, ok := ctx.Value(authKey{}).(*authCreds); ok {
		config.Username = u.Username
		config.Password = u.Password
	}

	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	prefix := DefaultPrefix
	f, ok := ctx.Value(prefixKey{}).(string)
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

func (c *etcd) String() string { return "etcd" }

func (c *etcd) Del(key string, opts ...kvstore.Option) error {
	ctx := context.TODO()
	for _, o := range opts {
		ctx = o(ctx)
	}

	ecopts := []clientv3.OpOption{}
	if val := ctx.Value(matchPrefix{}); val != nil {
		ecopts = append(ecopts, clientv3.WithPrefix())
	}
	_, err := c.client.Delete(ctx, c.prefix+key, ecopts...)
	return err
}

func (c *etcd) Set(key string, data []byte, opts ...kvstore.Option) error {
	ctx := context.TODO()
	for _, o := range opts {
		ctx = o(ctx)
	}

	_, err := c.client.Put(ctx, c.prefix+key, string(data))
	return err
}

func (c *etcd) Get(key string, opts ...kvstore.Option) ([]kvstore.KVPair, error) {
	ctx := context.TODO()
	for _, o := range opts {
		ctx = o(ctx)
	}

	ecopts := []clientv3.OpOption{}
	if val := ctx.Value(matchPrefix{}); val != nil {
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

func (c *etcd) Watch(opts ...kvstore.Option) kvstore.Watcher {
	ctx := context.TODO()
	for _, o := range opts {
		ctx = o(ctx)
	}

	key := ""
	if val := ctx.Value(watchKey{}); val != nil {
		key = val.(string)
	}

	ecopts := []clientv3.OpOption{clientv3.WithCreatedNotify()}
	if val := ctx.Value(matchPrefix{}); val != nil {
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
