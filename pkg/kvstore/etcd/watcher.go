package etcd

import (
	"context"
	"io"
	"strings"

	"github.com/cocktail828/go-tools/pkg/kvstore"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type watcher struct {
	ctx        context.Context
	cancel     context.CancelFunc
	prefix     string
	notifyChan clientv3.WatchChan
}

func (w *watcher) Next() ([]kvstore.Event, error) {
	select {
	case <-w.ctx.Done():
		return nil, io.EOF
	case val, ok := <-w.notifyChan:
		if !ok {
			return nil, io.EOF
		}

		events := []kvstore.Event{}
		for _, e := range val.Events {
			events = append(events, kvstore.Event{
				Type: kvstore.EventType(e.Type),
				Key:  strings.TrimPrefix(string(e.Kv.Key), w.prefix),
				Val:  e.Kv.Value,
			})
		}
		return events, nil
	}
}

func (w *watcher) Stop() error {
	w.cancel()
	return nil
}
