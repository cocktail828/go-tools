package environ

import (
	"log"
	"os"
	"regexp"
	"strconv"
)

var (
	re          = regexp.MustCompile(`(\w+)=([^\s]+)`)
	environVars = map[string]string{}
	mode        = PreLoad
)

type Mode int

const (
	PreLoad    Mode = iota // 性能优先, 预加载模式, 后续设置的环境变量不能获取到
	AlwaysLoad Mode = iota
)

func SetPolicy(m Mode) { mode = m }
func GetPolicy() Mode  { return mode }

func init() {
	for _, str := range os.Environ() {
		match := re.FindStringSubmatch(str)
		if len(match) > 0 {
			key, value := match[1], match[2]
			environVars[key] = value
		}
	}
}

func Getenv(name string) (string, bool) {
	if mode == PreLoad {
		val, ok := environVars[name]
		return val, ok
	}
	return os.LookupEnv(name)
}

func Exist(name string) bool {
	_, ok := Getenv(name)
	return ok
}

type option struct {
	bv      bool
	sv      string
	f32v    float32
	f64v    float64
	i64v    int64
	require bool
}

type Option func(*option)

// the env is required and must be set
func Required() Option             { return func(o *option) { o.require = true } }
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

func parseValue[T any](name string, parseFunc func(string) (T, error), defaultValue T, req bool) T {
	val, ok := Getenv(name)
	if !ok {
		if req {
			log.Fatalf("env '%q' is required but not found", name)
		}
		return defaultValue
	}

	if v, err := parseFunc(val); err == nil {
		return v
	}
	return defaultValue
}

func String(name string, opts ...Option) string {
	o := newOption(opts...)
	return parseValue(name, func(s string) (string, error) {
		return s, nil
	}, o.sv, o.require)
}

func Float32(name string, opts ...Option) float32 {
	o := newOption(opts...)
	return parseValue(name, func(s string) (float32, error) {
		v, err := strconv.ParseFloat(s, 32)
		return float32(v), err
	}, o.f32v, o.require)
}

func Float64(name string, opts ...Option) float64 {
	o := newOption(opts...)
	return parseValue(name, func(s string) (float64, error) {
		v, err := strconv.ParseFloat(s, 32)
		return v, err
	}, o.f64v, o.require)
}

func Int64(name string, opts ...Option) int64 {
	o := newOption(opts...)
	return parseValue(name, func(s string) (int64, error) {
		return strconv.ParseInt(s, 0, 64)
	}, o.i64v, o.require)
}

func Bool(name string, opts ...Option) bool {
	o := newOption(opts...)
	return parseValue(name, func(s string) (bool, error) {
		if v, err := strconv.ParseBool(s); err == nil {
			return v, nil
		}
		if v, err := strconv.ParseInt(s, 0, 64); err == nil {
			return v != 0, nil
		}
		return false, nil
	}, o.bv, o.require)
}
