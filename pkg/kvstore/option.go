package kvstore

import "github.com/cocktail828/go-tools/z/variadic"

type inVariadic struct{ variadic.Param }

func Variadic(opts ...variadic.Option) inVariadic {
	return inVariadic{variadic.Compose(opts...)}
}

type ttlKey struct{}

func TTL(val uint32) variadic.Option { return variadic.SetValue(ttlKey{}, val) }
func (iv inVariadic) TTL() uint32    { return variadic.GetValue[uint32](iv, ttlKey{}) }

type keepaliveKey struct{}

type KeepAliveCallback func(id, ttl int64, cancel func())

func KeepAlive(f KeepAliveCallback) variadic.Option { return variadic.SetValue(keepaliveKey{}, f) }
func (iv inVariadic) KeepAlive() KeepAliveCallback {
	return variadic.GetValue[KeepAliveCallback](iv, keepaliveKey{})
}

type prefixKey struct{}

func MatchPrefix() variadic.Option      { return variadic.SetValue(prefixKey{}, true) }
func (iv inVariadic) MatchPrefix() bool { return variadic.GetValue[bool](iv, prefixKey{}) }

type noLeaseKey struct{}

// ignore expired keys
func IgnoreLease() variadic.Option      { return variadic.SetValue(noLeaseKey{}, true) }
func (iv inVariadic) IgnoreLease() bool { return variadic.GetValue[bool](iv, noLeaseKey{}) }

type limitKey struct{}

// set the batch size of get
func Limit(val uint32) variadic.Option { return variadic.SetValue(limitKey{}, val) }
func (iv inVariadic) Limit() uint32    { return variadic.GetValue[uint32](iv, limitKey{}) }

type countKey struct{}

// get num of key-value pairs
func Count() variadic.Option      { return variadic.SetValue(countKey{}, true) }
func (iv inVariadic) Count() bool { return variadic.GetValue[bool](iv, countKey{}) }

type fromKey struct{}

// 分页查询开始key, 如果为空则从第一个开始
// nextKey = $(lastKey) + "\x00", 在 etcd 中，键是按字典序排序的。通过追加 \x00，可以确保下一个键是当前键的后一个键
func FromKey() variadic.Option      { return variadic.SetValue(fromKey{}, true) }
func (iv inVariadic) FromKey() bool { return variadic.GetValue[bool](iv, fromKey{}) }

type keyonlyKey struct{}

func KeyOnly() variadic.Option      { return variadic.SetValue(keyonlyKey{}, true) }
func (iv inVariadic) KeyOnly() bool { return variadic.GetValue[bool](iv, keyonlyKey{}) }
