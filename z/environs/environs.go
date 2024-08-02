package environs

import (
	"os"
	"strconv"
)

func Has(name string) bool {
	_, ok := os.LookupEnv(name)
	return ok
}

func String(name string) string {
	s, _ := os.LookupEnv(name)
	return s
}

func Float32(name string) float32 {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseFloat(val, 32); e == nil {
			return float32(v)
		}
	}
	return 0
}

func Float64(name string) float64 {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseFloat(val, 64); e == nil {
			return v
		}
	}
	return 0
}

func Int64(name string) int64 {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseInt(val, 0, 64); e == nil {
			return v
		}
	}
	return 0
}

func Bool(name string) bool {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseBool(val); e == nil {
			return v
		}
		if v, e := strconv.ParseInt(val, 0, 64); e == nil {
			return v != 0
		}
	}
	return false
}
