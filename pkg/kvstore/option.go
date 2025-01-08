package kvstore

import (
	"reflect"
)

type Option func(Param) Param
type Param interface {
	Value(key any) any
}

func WithValue(parent Param, key, val any) Param {
	if parent == nil {
		panic("cannot create Param from nil parent")
	}
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	return valueParam{parent, key, val}
}

// A valueParam carries a key-value pair. It implements Value for that key and
// delegates all other calls to the embedded Param.
type valueParam struct {
	Param
	key, val any
}

func (c valueParam) Value(key any) any {
	if c.key == key {
		return c.val
	}
	return c.Param.Value(key)
}

// implements...
// It is the common base of nopParam.
type nopParam struct{}

func (nopParam) Value(key any) any {
	return nil
}

// processing additional Parameters
type variadic struct{ Param }

func Variadic(opts ...Option) variadic {
	var p Param = nopParam{}
	for _, o := range opts {
		p = o(p)
	}
	return variadic{p}
}

func setValue(key, val any) Option {
	return func(parent Param) Param {
		return valueParam{parent, key, val}
	}
}

func getValue[T any](v variadic, key any) T {
	if val, ok := v.Value(key).(T); ok {
		return val
	}
	var zero T
	return zero
}

type ttlKey struct{}

func TTL(val uint32) Option    { return setValue(ttlKey{}, val) }
func (v variadic) TTL() uint32 { return getValue[uint32](v, ttlKey{}) }

type keepaliveKey struct{}

type KeepAliveCallback func(id, ttl int64, cancel func())

func KeepAlive(f KeepAliveCallback) Option { return setValue(keepaliveKey{}, f) }
func (v variadic) KeepAlive() KeepAliveCallback {
	return getValue[KeepAliveCallback](v, keepaliveKey{})
}

type prefixKey struct{}

func MatchPrefix() Option            { return setValue(prefixKey{}, true) }
func (v variadic) MatchPrefix() bool { return getValue[bool](v, prefixKey{}) }

type noLeaseKey struct{}

// ignore expired keys
func IgnoreLease() Option            { return setValue(noLeaseKey{}, true) }
func (v variadic) IgnoreLease() bool { return getValue[bool](v, noLeaseKey{}) }

type limitKey struct{}

// set the batch size of get
func Limit(val uint32) Option    { return setValue(limitKey{}, val) }
func (v variadic) Limit() uint32 { return getValue[uint32](v, limitKey{}) }

type countKey struct{}

// get num of key-value pairs
func Count() Option            { return setValue(countKey{}, true) }
func (v variadic) Count() bool { return getValue[bool](v, countKey{}) }

type fromKey struct{}

// 分页查询开始key, 如果为空则从第一个开始
// nextKey = $(lastKey) + "\x00", 在 etcd 中，键是按字典序排序的。通过追加 \x00，可以确保下一个键是当前键的后一个键
func FromKey() Option            { return setValue(fromKey{}, true) }
func (v variadic) FromKey() bool { return getValue[bool](v, fromKey{}) }

type keyonlyKey struct{}

func KeyOnly() Option            { return setValue(keyonlyKey{}, true) }
func (v variadic) KeyOnly() bool { return getValue[bool](v, keyonlyKey{}) }
