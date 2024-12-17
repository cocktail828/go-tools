package environs

import (
	"os"
	"regexp"
	"strconv"
)

var (
	re          = regexp.MustCompile(`(\w+)=([^\s]+)`)
	environVars = map[string]string{}
)

func init() {
	for _, str := range os.Environ() {
		match := re.FindStringSubmatch(str)
		if len(match) > 0 {
			key, value := match[1], match[2]
			environVars[key] = value
		}
	}
}

func Has(name string) bool {
	_, ok := environVars[name]
	return ok
}

type option struct {
	bv   bool
	sv   string
	f32v float32
	f64v float64
	i64v int64
}

type Option func(*option)

func WithBool(v bool) Option       { return func(o *option) { o.bv = v } }
func WithString(v string) Option   { return func(o *option) { o.sv = v } }
func WithFloat32(v float32) Option { return func(o *option) { o.f32v = v } }
func WithFloat64(v float64) Option { return func(o *option) { o.f64v = v } }
func WithInt64(v int64) Option     { return func(o *option) { o.i64v = v } }

func newOption(opts ...Option) option {
	o := option{}
	for _, f := range opts {
		f(&o)
	}
	return o
}

func String(name string, opts ...Option) string {
	o := newOption(opts...)
	if !Has(name) {
		return o.sv
	}

	return environVars[name]
}

func Float32(name string, opts ...Option) float32 {
	o := newOption(opts...)
	if !Has(name) {
		return o.f32v
	}

	if val, ok := environVars[name]; ok {
		if v, e := strconv.ParseFloat(val, 32); e == nil {
			return float32(v)
		}
	}
	return 0
}

func Float64(name string, opts ...Option) float64 {
	o := newOption(opts...)
	if !Has(name) {
		return o.f64v
	}

	if val, ok := environVars[name]; ok {
		if v, e := strconv.ParseFloat(val, 64); e == nil {
			return v
		}
	}
	return 0
}

func Int64(name string, opts ...Option) int64 {
	o := newOption(opts...)
	if !Has(name) {
		return o.i64v
	}

	if val, ok := environVars[name]; ok {
		if v, e := strconv.ParseInt(val, 0, 64); e == nil {
			return v
		}
	}
	return 0
}

func Bool(name string, opts ...Option) bool {
	o := newOption(opts...)
	if !Has(name) {
		return o.bv
	}

	if val, ok := environVars[name]; ok {
		if v, e := strconv.ParseBool(val); e == nil {
			return v
		}
		if v, e := strconv.ParseInt(val, 0, 64); e == nil {
			return v != 0
		}
	}
	return false
}
