package environ

import (
	"os"
	"strconv"
	"strings"

	"github.com/cocktail828/go-tools/z/variadic"
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

type inVariadic struct{ variadic.Assigned }

type boolKey struct{}

func WithBool(v bool) variadic.Option { return variadic.SetValue(boolKey{}, v) }
func (iv inVariadic) WithBool() bool  { return variadic.GetValue[bool](iv, boolKey{}) }

type stringKey struct{}

func WithString(v string) variadic.Option { return variadic.SetValue(stringKey{}, v) }
func (iv inVariadic) WithString() string  { return variadic.GetValue[string](iv, stringKey{}) }

type float64Key struct{}

func WithFloat64(v float64) variadic.Option { return variadic.SetValue(float64Key{}, v) }
func (iv inVariadic) WithFloat64() float64  { return variadic.GetValue[float64](iv, float64Key{}) }

type int64Key struct{}

func WithInt64(v int64) variadic.Option { return variadic.SetValue(int64Key{}, v) }
func (iv inVariadic) WithInt64() int64  { return variadic.GetValue[int64](iv, int64Key{}) }

func parseValue[T any](name string, parseFunc func(string) (T, error), defaultValue T) T {
	val, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	if v, err := parseFunc(val); err == nil {
		return v
	}
	return defaultValue
}

func String(name string, opts ...variadic.Option) string {
	iv := inVariadic{variadic.Compose(opts...)}
	return parseValue(name, func(s string) (string, error) {
		return s, nil
	}, iv.WithString())
}

func Float64(name string, opts ...variadic.Option) float64 {
	iv := inVariadic{variadic.Compose(opts...)}
	return parseValue(name, func(s string) (float64, error) {
		v, err := strconv.ParseFloat(s, 64)
		return v, err
	}, iv.WithFloat64())
}

func Int64(name string, opts ...variadic.Option) int64 {
	iv := inVariadic{variadic.Compose(opts...)}
	return parseValue(name, func(s string) (int64, error) {
		return strconv.ParseInt(s, 0, 64)
	}, iv.WithInt64())
}

// load bool env loosely, accept "true", "false", "0", and non-zero digits
func Bool(name string, opts ...variadic.Option) bool {
	iv := inVariadic{variadic.Compose(opts...)}
	return parseValue(name, func(s string) (bool, error) {
		if v, err := strconv.ParseBool(s); err == nil {
			return v, nil
		}
		if v, err := strconv.ParseInt(s, 0, 64); err == nil {
			return v != 0, nil
		}
		return false, nil
	}, iv.WithBool())
}
