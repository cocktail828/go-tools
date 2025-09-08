package environ

import (
	"os"
	"strconv"
	"strings"
)

func Lookup() map[string]string {
	envs := map[string]string{}
	for _, str := range os.Environ() {
		parts := strings.SplitN(str, "=", 2)
		if len(parts) == 2 {
			envs[parts[0]] = parts[1]
		}
	}
	return envs
}

type option struct {
	boolVal  bool
	strVal   string
	floatVal float64
	intVal   int64
}

func compose(opts ...Option) *option {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type Option func(o *option)

func WithBool(v bool) Option     { return func(o *option) { o.boolVal = v } }
func WithString(v string) Option { return func(o *option) { o.strVal = v } }
func WithFloat(v float64) Option { return func(o *option) { o.floatVal = v } }
func WithInt(v int64) Option     { return func(o *option) { o.intVal = v } }

func parseValue[T any](name string, parseFunc func(string) (T, error), defaultValue T) T {
	if name == "" || name == "-" {
		return defaultValue
	}
	val, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	if v, err := parseFunc(val); err == nil {
		return v
	}
	return defaultValue
}

func String(name string, opts ...Option) string {
	return parseValue(name, func(s string) (string, error) {
		return s, nil
	}, compose(opts...).strVal)
}

func Float(name string, opts ...Option) float64 {
	return parseValue(name, func(s string) (float64, error) {
		v, err := strconv.ParseFloat(s, 64)
		return v, err
	}, compose(opts...).floatVal)
}

func Int(name string, opts ...Option) int64 {
	return parseValue(name, func(s string) (int64, error) {
		return strconv.ParseInt(s, 0, 64)
	}, compose(opts...).intVal)
}

// load bool env loosely, accept "true", "false", "0", and non-zero digits
func Bool(name string, opts ...Option) bool {
	return parseValue(name, strconv.ParseBool, compose(opts...).boolVal)
}
