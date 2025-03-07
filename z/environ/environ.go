package environ

import (
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/cocktail828/go-tools/z/variadic"
)

var (
	re      = regexp.MustCompile(`(\w+)=([^\s]+)`)
	envVars = map[string]string{}
)

func init() {
	for _, str := range os.Environ() {
		match := re.FindStringSubmatch(str)
		if len(match) > 0 {
			key, value := match[1], match[2]
			envVars[key] = value
		}
	}
}

func Getenv(name string) (string, bool) {
	if val, ok := envVars[name]; ok {
		return val, ok
	}
	return os.LookupEnv(name)
}

func Exist(name string) bool {
	_, ok := Getenv(name)
	return ok
}

type inVariadic struct{ variadic.Assigned }

type reqKey struct{}

func Required() variadic.Option      { return variadic.SetValue(reqKey{}, true) }
func (iv inVariadic) Required() bool { return variadic.GetValue[bool](iv, reqKey{}) }

type boolKey struct{}

func WithBool(v bool) variadic.Option { return variadic.SetValue(boolKey{}, v) }
func (iv inVariadic) WithBool() bool  { return variadic.GetValue[bool](iv, boolKey{}) }

type stringKey struct{}

func WithString(v string) variadic.Option { return variadic.SetValue(stringKey{}, v) }
func (iv inVariadic) WithString() string  { return variadic.GetValue[string](iv, stringKey{}) }

type float32Key struct{}

func WithFloat32(v float32) variadic.Option { return variadic.SetValue(float32Key{}, v) }
func (iv inVariadic) WithFloat32() float32  { return variadic.GetValue[float32](iv, float32Key{}) }

type float64Key struct{}

func WithFloat64(v float64) variadic.Option { return variadic.SetValue(float64Key{}, v) }
func (iv inVariadic) WithFloat64() float64  { return variadic.GetValue[float64](iv, float64Key{}) }

type int64Key struct{}

func WithInt64(v int64) variadic.Option { return variadic.SetValue(int64Key{}, v) }
func (iv inVariadic) WithInt64() int64  { return variadic.GetValue[int64](iv, int64Key{}) }

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

func String(name string, opts ...variadic.Option) string {
	iv := inVariadic{variadic.Compose(opts...)}
	return parseValue(name, func(s string) (string, error) {
		return s, nil
	}, iv.WithString(), iv.Required())
}

func Float32(name string, opts ...variadic.Option) float32 {
	iv := inVariadic{variadic.Compose(opts...)}
	return parseValue(name, func(s string) (float32, error) {
		v, err := strconv.ParseFloat(s, 32)
		return float32(v), err
	}, iv.WithFloat32(), iv.Required())
}

func Float64(name string, opts ...variadic.Option) float64 {
	iv := inVariadic{variadic.Compose(opts...)}
	return parseValue(name, func(s string) (float64, error) {
		v, err := strconv.ParseFloat(s, 32)
		return v, err
	}, iv.WithFloat64(), iv.Required())
}

func Int64(name string, opts ...variadic.Option) int64 {
	iv := inVariadic{variadic.Compose(opts...)}
	return parseValue(name, func(s string) (int64, error) {
		return strconv.ParseInt(s, 0, 64)
	}, iv.WithInt64(), iv.Required())
}

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
	}, iv.WithBool(), iv.Required())
}
