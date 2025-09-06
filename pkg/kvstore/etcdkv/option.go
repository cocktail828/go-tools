package etcdkv

import (
	"github.com/cocktail828/go-tools/pkg/kvstore"
)

type CountResult struct{ Num int }

func (r CountResult) Len() int         { return r.Num }
func (r CountResult) Key(int) string   { return "" }
func (r CountResult) Value(int) []byte { return nil }

var _ kvstore.SetOption = (*etcdSetOption)(nil)

type etcdSetOption struct {
	ttl       int64 // in second
	keepalive bool
}

func newEtcdSetOption(opts ...kvstore.SetOption) *etcdSetOption {
	setopt := &etcdSetOption{}
	for _, o := range opts {
		if f, ok := o.(func(*etcdSetOption)); ok {
			f(setopt)
		}
	}
	return setopt
}

func WithTTL(ttl int64) kvstore.SetOption     { return func(o *etcdSetOption) { o.ttl = ttl } }
func WithKeepAlive(ka bool) kvstore.SetOption { return func(o *etcdSetOption) { o.keepalive = ka } }

var _ kvstore.GetOption = (*etcdGetOption)(nil)

type etcdGetOption struct {
	matchprefix bool
	count       bool
	ignorelease bool
	keyonly     bool
	limit       int
	fromKey     bool
}

func newEtcdGetOption(opts ...kvstore.GetOption) *etcdGetOption {
	getopt := &etcdGetOption{}
	for _, o := range opts {
		if f, ok := o.(func(*etcdGetOption)); ok {
			f(getopt)
		}
	}
	return getopt
}

func WithMatchPrefix() kvstore.GetOption    { return func(o *etcdGetOption) { o.matchprefix = true } }
func WithCount() kvstore.GetOption          { return func(o *etcdGetOption) { o.count = true } }
func WithIgnoreLease() kvstore.GetOption    { return func(o *etcdGetOption) { o.ignorelease = true } }
func WithKeyOnly() kvstore.GetOption        { return func(o *etcdGetOption) { o.keyonly = true } }
func WithLimit(limit int) kvstore.GetOption { return func(o *etcdGetOption) { o.limit = limit } }
func WithFromKey() kvstore.GetOption        { return func(o *etcdGetOption) { o.fromKey = true } }

var _ kvstore.DelOption = (*etcdDelOption)(nil)

type etcdDelOption struct {
	prefix string
}

func newEtcdDelOption(opts ...kvstore.DelOption) *etcdDelOption {
	delopt := &etcdDelOption{}
	for _, o := range opts {
		if f, ok := o.(func(*etcdDelOption)); ok {
			f(delopt)
		}
	}
	return delopt
}

func DelWithPrefix(prefix string) kvstore.DelOption {
	return func(o *etcdDelOption) { o.prefix = prefix }
}

var _ kvstore.WatchOption = (*etcdWatchOption)(nil)

type etcdWatchOption struct {
	prefix string
}

func newEtcdWatchOption(opts ...kvstore.WatchOption) *etcdWatchOption {
	watchopt := &etcdWatchOption{}
	for _, o := range opts {
		if f, ok := o.(func(*etcdWatchOption)); ok {
			f(watchopt)
		}
	}
	return watchopt
}

func WatchWithPrefix(prefix string) kvstore.WatchOption {
	return func(o *etcdWatchOption) { o.prefix = prefix }
}
