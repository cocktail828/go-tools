package variadic

import "reflect"

type Option func(Assigned) Assigned
type Assigned interface {
	Value(key any) any
}

func WithValue(parent Assigned, key, val any) Assigned {
	if parent == nil {
		panic("cannot create Assigned from nil parent")
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
// delegates all other calls to the embedded Assigned.
type valueParam struct {
	Assigned
	key, val any
}

func (c valueParam) Value(key any) any {
	if c.key == key {
		return c.val
	}
	return c.Assigned.Value(key)
}

// implements...
// It is the common base of nopParam.
type nopParam struct{}

func (nopParam) Value(key any) any {
	return nil
}

func Compose(opts ...Option) Assigned {
	var p Assigned = nopParam{}
	for _, o := range opts {
		p = o(p)
	}
	return p
}

func SetValue(key, val any) Option {
	return func(parent Assigned) Assigned {
		return valueParam{parent, key, val}
	}
}

func GetValue[T any](p Assigned, key any) T {
	if val, ok := p.Value(key).(T); ok {
		return val
	}
	var zero T
	return zero
}
