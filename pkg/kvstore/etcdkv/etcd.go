package etcdkv

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/cocktail828/go-tools/pkg/kvstore"
	"github.com/cocktail828/go-tools/pkg/kvstore/common"
	"github.com/cocktail828/go-tools/z/environ"
	"github.com/cocktail828/go-tools/z/variadic"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdKV struct {
	root   string
	cfg    clientv3.Config
	client *clientv3.Client
}

func New(cfg clientv3.Config, root string) (kvstore.KV, error) {
	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &etcdKV{client: client, root: root}, nil
}

func (e *etcdKV) String() string {
	u := url.URL{
		Scheme: "etcd",
		User:   url.UserPassword(e.cfg.Username, e.cfg.Password),
		Host:   strings.Join(e.cfg.Endpoints, ","),
	}
	return u.String()
}

func (e *etcdKV) Set(ctx context.Context, key string, val []byte, opts ...variadic.Option) (err error) {
	options := []clientv3.OpOption{}
	v := kvstore.Variadic(opts...)

	var lease *clientv3.LeaseGrantResponse
	if val := v.TTL(); val > 0 {
		if lease, err = e.client.Grant(ctx, int64(val)); err != nil {
			return err
		}
	}

	if f := v.KeepAlive(); f != nil {
		if lease == nil {
			ttl := environ.Int64("ETCDKV_KEEPALIVE_TTL", environ.WithInt64(5))
			if lease, err = e.client.Grant(ctx, ttl); err != nil {
				return err
			}
		}

		kaCtx, kaCancel := context.WithCancel(context.TODO())
		keepAliveCh, err := e.client.KeepAlive(kaCtx, lease.ID)
		if err != nil {
			kaCancel()
			return err
		}

		go func() {
			defer kaCancel()
			if f == nil {
				f = func(id, ttl int64, cancel func()) {}
			}

			for {
				select {
				case resp := <-keepAliveCh:
					if resp == nil {
						f(int64(lease.ID), 0, kaCancel)
						return
					}
					f(int64(lease.ID), lease.TTL, kaCancel)
				case <-kaCtx.Done():
					f(int64(lease.ID), 0, kaCancel)
					return
				}
			}
		}()
	}

	if lease != nil {
		options = append(options, clientv3.WithLease(lease.ID))
	}

	_, err = e.client.Put(ctx, path.Join(e.root, key), string(val), options...)
	return err
}

func (e *etcdKV) Get(ctx context.Context, key string, opts ...variadic.Option) (kvstore.Result, error) {
	options := []clientv3.OpOption{}
	v := kvstore.Variadic(opts...)
	if v.MatchPrefix() {
		options = append(options, clientv3.WithPrefix())
	}

	if v.IgnoreLease() {
		options = append(options, clientv3.WithIgnoreLease())
	}

	isCount := false
	if v.Count() || v.KeyOnly() {
		isCount = true
		options = append(options, clientv3.WithKeysOnly())
	}

	if val := v.Limit(); val > 0 {
		options = append(options, clientv3.WithLimit(int64(val)))
	}

	if v.FromKey() {
		options = append(options, clientv3.WithFromKey())
	}

	ev, err := e.client.Get(ctx, path.Join(e.root, key), options...)
	if err != nil {
		return nil, err
	}

	result := common.Result{}
	for _, kv := range ev.Kvs {
		result.Append(e.normlizeKey(kv.Key), kv.Value)
	}

	if isCount {
		return CountResult{Num: result.Len()}, nil
	}

	return result, nil
}

func (e *etcdKV) normlizeKey(key []byte) string {
	return strings.TrimPrefix(string(key), e.root+"/")
}

func (e *etcdKV) Del(ctx context.Context, key string, opts ...variadic.Option) error {
	options := []clientv3.OpOption{}
	v := kvstore.Variadic(opts...)
	if v.MatchPrefix() {
		options = append(options, clientv3.WithPrefix())
	}

	_, err := e.client.Delete(ctx, path.Join(e.root, key), options...)
	return err
}

func (e *etcdKV) Watch(ctx context.Context, opts ...variadic.Option) kvstore.Watcher {
	options := []clientv3.OpOption{}
	v := kvstore.Variadic(opts...)
	if v.MatchPrefix() {
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

func (w *etcdWatcher) Next(ctx context.Context) (kvstore.Event, error) {
	select {
	case etcdEvent, ok := <-w.watchChan:
		if !ok {
			return nil, kvstore.ErrWatcherStopped
		}

		event := common.Event{}
		for _, ev := range etcdEvent.Events {
			tp := kvstore.NONE
			switch ev.Type {
			case clientv3.EventTypePut:
				tp = kvstore.PUT
			case clientv3.EventTypeDelete:
				tp = kvstore.DELETE
			}

			event.Append(tp, w.kv.normlizeKey(ev.Kv.Key), ev.Kv.Value)
		}
		return event, nil

	case <-w.runningCtx.Done():
		return nil, kvstore.ErrWatcherStopped

	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

func (w *etcdWatcher) Stop() error {
	w.runningCacnel()
	return nil
}
