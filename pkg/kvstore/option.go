package kvstore

import "github.com/cocktail828/go-tools/z/variadic"

type ttlKey struct{}

func TTL(val uint32) variadic.Option     { return variadic.Set(ttlKey{}, val) }
func GetTTL(c variadic.Container) uint32 { return variadic.Value[uint32](c, ttlKey{}) }

type keepaliveKey struct{}

type KeepAliveCallback func(id, ttl int64, cancel func())

func KeepAlive(f KeepAliveCallback) variadic.Option { return variadic.Set(keepaliveKey{}, f) }
func GetKeepAlive(c variadic.Container) KeepAliveCallback {
	return variadic.Value[KeepAliveCallback](c, keepaliveKey{})
}

type prefixKey struct{}

func MatchPrefix() variadic.Option             { return variadic.Set(prefixKey{}, true) }
func GetMatchPrefix(c variadic.Container) bool { return variadic.Value[bool](c, prefixKey{}) }

type noLeaseKey struct{}

// ignore expired keys
func IgnoreLease() variadic.Option             { return variadic.Set(noLeaseKey{}, true) }
func GetIgnoreLease(c variadic.Container) bool { return variadic.Value[bool](c, noLeaseKey{}) }

type limitKey struct{}

// set the batch size of get
func Limit(val uint32) variadic.Option     { return variadic.Set(limitKey{}, val) }
func GetLimit(c variadic.Container) uint32 { return variadic.Value[uint32](c, limitKey{}) }

type countKey struct{}

// get num of key-value pairs
func Count() variadic.Option             { return variadic.Set(countKey{}, true) }
func GetCount(c variadic.Container) bool { return variadic.Value[bool](c, countKey{}) }

type fromKey struct{}

// 分页查询开始key, 如果为空则从第一个开始
// nextKey = $(lastKey) + "\x00", 在 etcd 中，键是按字典序排序的。通过追加 \x00，可以确保下一个键是当前键的后一个键
func FromKey() variadic.Option             { return variadic.Set(fromKey{}, true) }
func GetFromKey(c variadic.Container) bool { return variadic.Value[bool](c, fromKey{}) }

type keyonlyKey struct{}

func KeyOnly() variadic.Option             { return variadic.Set(keyonlyKey{}, true) }
func GetKeyOnly(c variadic.Container) bool { return variadic.Value[bool](c, keyonlyKey{}) }
