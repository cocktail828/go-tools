package etcdkv

import "github.com/cocktail828/go-tools/pkg/kv"

type CountResult struct{ Num int }

func (r CountResult) Len() int         { return r.Num }
func (r CountResult) Key(int) string   { return "" }
func (r CountResult) Value(int) []byte { return nil }

type etcdSetOption struct {
	ttl       int64 // in second
	keepalive bool
}

func newEtcdSetOption(opts ...kv.SetOption) *etcdSetOption {
	setopt := &etcdSetOption{}
	for _, o := range opts {
		o(setopt)
	}
	return setopt
}

func WithTTL(ttl int64) kv.SetOption {
	return func(o any) {
		if opt, ok := o.(*etcdSetOption); ok {
			opt.ttl = ttl
		}
	}
}

func WithKeepAlive(ka bool) kv.SetOption {
	return func(o any) {
		if opt, ok := o.(*etcdSetOption); ok {
			opt.keepalive = ka
		}
	}
}

func WithMatchPrefix() kv.GetOption {
	return func(o any) {
		if opt, ok := o.(*etcdGetOption); ok {
			opt.matchprefix = true
		}
	}
}

type etcdGetOption struct {
	matchprefix bool
	count       bool
	ignorelease bool
	keyonly     bool
	limit       int
	fromKey     bool
}

func newEtcdGetOption(opts ...kv.GetOption) *etcdGetOption {
	getopt := &etcdGetOption{}
	for _, o := range opts {
		o(getopt)
	}
	return getopt
}

func WithCount() kv.GetOption {
	return func(o any) {
		if opt, ok := o.(*etcdGetOption); ok {
			opt.count = true
		}
	}
}
func WithIgnoreLease() kv.GetOption {
	return func(o any) {
		if opt, ok := o.(*etcdGetOption); ok {
			opt.ignorelease = true
		}
	}
}
func WithKeyOnly() kv.GetOption {
	return func(o any) {
		if opt, ok := o.(*etcdGetOption); ok {
			opt.keyonly = true
		}
	}
}
func WithLimit(limit int) kv.GetOption {
	return func(o any) {
		if opt, ok := o.(*etcdGetOption); ok {
			opt.limit = limit
		}
	}
}
func WithFromKey() kv.GetOption {
	return func(o any) {
		if opt, ok := o.(*etcdGetOption); ok {
			opt.fromKey = true
		}
	}
}

type etcdDelOption struct {
	matchprefix bool
}

func newEtcdDelOption(opts ...kv.DelOption) *etcdDelOption {
	delopt := &etcdDelOption{}
	for _, o := range opts {
		o(delopt)
	}
	return delopt
}

func DelWithPrefix() kv.DelOption {
	return func(o any) {
		if opt, ok := o.(*etcdDelOption); ok {
			opt.matchprefix = true
		}
	}
}

type etcdWatchOption struct {
	matchprefix bool
}

func newEtcdWatchOption(opts ...kv.WatchOption) *etcdWatchOption {
	watchopt := &etcdWatchOption{}
	for _, o := range opts {
		o(watchopt)
	}
	return watchopt
}

func WatchWithPrefix() kv.WatchOption {
	return func(o any) {
		if opt, ok := o.(*etcdWatchOption); ok {
			opt.matchprefix = true
		}
	}
}
