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

type boolKey struct{}

func WithBool(v bool) variadic.Option   { return variadic.Set(boolKey{}, v) }
func getBool(c variadic.Container) bool { return variadic.Value[bool](c, boolKey{}) }

type stringKey struct{}

func WithString(v string) variadic.Option   { return variadic.Set(stringKey{}, v) }
func getString(c variadic.Container) string { return variadic.Value[string](c, stringKey{}) }

type float64Key struct{}

func WithFloat64(v float64) variadic.Option   { return variadic.Set(float64Key{}, v) }
func getFloat64(c variadic.Container) float64 { return variadic.Value[float64](c, float64Key{}) }

type int64Key struct{}

func WithInt64(v int64) variadic.Option   { return variadic.Set(int64Key{}, v) }
func getInt64(c variadic.Container) int64 { return variadic.Value[int64](c, int64Key{}) }

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

func String(name string, opts ...variadic.Option) string {
	iv := variadic.Compose(opts...)
	return parseValue(name, func(s string) (string, error) {
		return s, nil
	}, getString(iv))
}

func Float64(name string, opts ...variadic.Option) float64 {
	iv := variadic.Compose(opts...)
	return parseValue(name, func(s string) (float64, error) {
		v, err := strconv.ParseFloat(s, 64)
		return v, err
	}, getFloat64(iv))
}

func Int64(name string, opts ...variadic.Option) int64 {
	iv := variadic.Compose(opts...)
	return parseValue(name, func(s string) (int64, error) {
		return strconv.ParseInt(s, 0, 64)
	}, getInt64(iv))
}

// load bool env loosely, accept "true", "false", "0", and non-zero digits
func Bool(name string, opts ...variadic.Option) bool {
	iv := variadic.Compose(opts...)
	return parseValue(name, strconv.ParseBool, getBool(iv))
}
