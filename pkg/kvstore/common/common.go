package common

import (
	"context"

	"github.com/cocktail828/go-tools/pkg/kvstore"
)

// Result 是 Result 和 Event 的公共基础结构体
type Result struct {
	Keys   []string
	Values [][]byte
}

// Append 向基础结构体中添加一个键值对
func (r *Result) Append(key string, val []byte) {
	r.Keys = append(r.Keys, key)
	r.Values = append(r.Values, val)
}

// Len 返回基础结构体中键值对的数量
func (r Result) Len() int {
	return len(r.Keys)
}

// Key 返回指定索引的键
func (r Result) Key(i int) string {
	if i < 0 || i >= r.Len() {
		return ""
	}
	return r.Keys[i]
}

// Value 返回指定索引的值
func (r Result) Value(i int) []byte {
	if i < 0 || i >= r.Len() {
		return nil
	}
	return r.Values[i]
}

// Event 实现了 kvstore.Event 接口，用于存储键值对事件
type Event struct {
	Result
	Types []kvstore.Type // 事件类型（PUT 或 DELETE）
}

// Append 向事件中添加一个键值对及其类型
func (e *Event) Append(eventType kvstore.Type, key string, val []byte) {
	e.Result.Append(key, val)
	e.Types = append(e.Types, eventType)
}

// Type 返回指定索引的事件类型
func (e Event) Type(i int) kvstore.Type {
	if i < 0 || i >= e.Len() {
		return kvstore.NONE // 默认值
	}
	return e.Types[i]
}

// NopWatcher 是一个空实现的 Watcher
type NopWatcher struct{}

// Next 返回未实现的错误
func (NopWatcher) Next(context.Context) (kvstore.Event, error) {
	return nil, kvstore.ErrNotImplement
}

// Stop 停止 Watcher，空实现
func (NopWatcher) Stop() error {
	return nil
}

var (
	_ kvstore.Result  = (*Result)(nil)
	_ kvstore.Event   = (*Event)(nil)
	_ kvstore.Watcher = NopWatcher{}
)
