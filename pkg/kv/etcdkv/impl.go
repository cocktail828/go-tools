package etcdkv

import (
	"github.com/cocktail828/go-tools/pkg/kv"
)

type etcdKvPairs struct {
	Keys   []string
	Values [][]byte
}

func (r *etcdKvPairs) Append(key string, val []byte) {
	r.Keys = append(r.Keys, key)
	r.Values = append(r.Values, val)
}

func (r etcdKvPairs) Len() int {
	return len(r.Keys)
}

func (r etcdKvPairs) Key(i int) string {
	if i < 0 || i >= r.Len() {
		return ""
	}
	return r.Keys[i]
}

func (r etcdKvPairs) Value(i int) []byte {
	if i < 0 || i >= r.Len() {
		return nil
	}
	return r.Values[i]
}

type etcdEvent struct {
	etcdKvPairs
	Types []kv.Type // 事件类型（PUT 或 DELETE）
}

func (e *etcdEvent) Append(eventType kv.Type, key string, val []byte) {
	e.etcdKvPairs.Append(key, val)
	e.Types = append(e.Types, eventType)
}

func (e etcdEvent) Type(i int) kv.Type {
	if i < 0 || i >= e.Len() {
		return kv.NONE
	}
	return e.Types[i]
}

var (
	_ kv.Result = (*etcdKvPairs)(nil)
	_ kv.Event  = (*etcdEvent)(nil)
)
