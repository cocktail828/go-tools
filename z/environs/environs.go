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
			key := match[1]
			value := match[2]
			environVars[key] = value
		}
	}
}

func Has(name string) bool {
	_, ok := environVars[name]
	return ok
}

func String(name string) string {
	return environVars[name]
}

func StringWithValue(name string, val string) string {
	if Has(name) {
		return String(name)
	}
	return val
}

func Float32(name string) float32 {
	if val, ok := environVars[name]; ok {
		if v, e := strconv.ParseFloat(val, 32); e == nil {
			return float32(v)
		}
	}
	return 0
}

func Float32WithValue(name string, val float32) float32 {
	if Has(name) {
		return Float32(name)
	}
	return val
}

func Float64(name string) float64 {
	if val, ok := environVars[name]; ok {
		if v, e := strconv.ParseFloat(val, 64); e == nil {
			return v
		}
	}
	return 0
}

func Float64WithValue(name string, val float64) float64 {
	if Has(name) {
		return Float64(name)
	}
	return val
}

func Int64(name string) int64 {
	if val, ok := environVars[name]; ok {
		if v, e := strconv.ParseInt(val, 0, 64); e == nil {
			return v
		}
	}
	return 0
}

func Int64WithValue(name string, val int64) int64 {
	if Has(name) {
		return Int64(name)
	}
	return val
}

func Bool(name string) bool {
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

func BoolWithValue(name string, val bool) bool {
	if Has(name) {
		return Bool(name)
	}
	return val
}
