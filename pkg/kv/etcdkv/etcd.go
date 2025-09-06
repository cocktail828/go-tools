package etcdkv

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/cocktail828/go-tools/pkg/kv"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdKV struct {
	root   string
	cfg    clientv3.Config
	client *clientv3.Client
}

func New(cfg clientv3.Config, root string) (kv.KV, error) {
	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &etcdKV{
		root:   root,
		cfg:    cfg,
		client: client,
	}, nil
}

func (e *etcdKV) String() string {
	u := url.URL{
		Scheme: "etcd",
		User:   url.UserPassword(e.cfg.Username, e.cfg.Password),
		Host:   strings.Join(e.cfg.Endpoints, ","),
	}
	return u.String()
}

func (e *etcdKV) Set(ctx context.Context, key string, val []byte, opts ...kv.SetOption) error {
	setopt := newEtcdSetOption(opts...)
	options := []clientv3.OpOption{}
	var lease *clientv3.LeaseGrantResponse
	var err error

	if setopt.ttl > 0 {
		lease, err = e.client.Grant(ctx, int64(setopt.ttl))
		if err != nil {
			return err
		}
		options = append(options, clientv3.WithLease(lease.ID))
	}

	_, err = e.client.Put(ctx, path.Join(e.root, key), string(val), options...)
	if err != nil {
		return err
	}

	if setopt.keepalive && lease != nil {
		e.client.KeepAlive(ctx, lease.ID)
	}

	return nil
}

func (e *etcdKV) Get(ctx context.Context, key string, opts ...kv.GetOption) (kv.Result, error) {
	getopt := newEtcdGetOption(opts...)
	options := []clientv3.OpOption{}

	if getopt.matchprefix {
		options = append(options, clientv3.WithPrefix())
	}

	if getopt.ignorelease {
		options = append(options, clientv3.WithIgnoreLease())
	}

	isCount := false
	if getopt.count || getopt.keyonly {
		isCount = true
		options = append(options, clientv3.WithKeysOnly())
	}

	if getopt.limit > 0 {
		options = append(options, clientv3.WithLimit(int64(getopt.limit)))
	}

	if getopt.fromKey {
		options = append(options, clientv3.WithFromKey())
	}

	ev, err := e.client.Get(ctx, path.Join(e.root, key), options...)
	if err != nil {
		return nil, err
	}

	pairs := etcdKvPairs{}
	for _, kv := range ev.Kvs {
		pairs.Append(e.normlizeKey(kv.Key), kv.Value)
	}

	if isCount {
		return CountResult{Num: pairs.Len()}, nil
	}

	return pairs, nil
}

func (e *etcdKV) normlizeKey(key []byte) string {
	return strings.TrimPrefix(string(key), e.root+"/")
}

func (e *etcdKV) Del(ctx context.Context, key string, opts ...kv.DelOption) error {
	options := []clientv3.OpOption{}
	delopt := newEtcdDelOption(opts...)
	if delopt.matchprefix {
		options = append(options, clientv3.WithPrefix())
	}

	_, err := e.client.Delete(ctx, path.Join(e.root, key), options...)
	return err
}

func (e *etcdKV) Watch(ctx context.Context, opts ...kv.WatchOption) kv.Watcher {
	options := []clientv3.OpOption{}
	watchopt := newEtcdWatchOption(opts...)
	if watchopt.matchprefix {
		options = append(options, clientv3.WithPrefix())
	}

	ctx, cancel := context.WithCancel(ctx)
	watchChan := e.client.Watch(ctx, e.root, options...)
	return &etcdWatcher{
		kv:            e,
		runningCtx:    ctx,
		runningCacnel: cancel,
		watchChan:     watchChan,
	}
}

func (e *etcdKV) Close() error {
	return e.client.Close()
}

type etcdWatcher struct {
	kv            *etcdKV
	runningCtx    context.Context
	runningCacnel context.CancelFunc
	watchChan     clientv3.WatchChan
}

func (w *etcdWatcher) Next(ctx context.Context) (kv.Event, error) {
	select {
	case event, ok := <-w.watchChan:
		if !ok {
			return nil, kv.ErrWatcherStopped
		}

		ev := &etcdEvent{}
		for _, e := range event.Events {
			tp := kv.NONE
			switch e.Type {
			case clientv3.EventTypePut:
				tp = kv.PUT
			case clientv3.EventTypeDelete:
				tp = kv.DELETE
			}
			ev.Append(tp, w.kv.normlizeKey(e.Kv.Key), e.Kv.Value)
		}
		return ev, nil

	case <-w.runningCtx.Done():
		return nil, kv.ErrWatcherStopped

	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

func (w *etcdWatcher) Stop() error {
	w.runningCacnel()
	return nil
}
