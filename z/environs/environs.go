package environs

import (
	"os"
	"strconv"
)

func String(name string) (string, bool) {
	return os.LookupEnv(name)
}

func Float32(name string) (float32, bool) {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseFloat(val, 32); e == nil {
			return float32(v), true
		}
	}
	return 0, false
}

func Float64(name string) (float64, bool) {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseFloat(val, 64); e == nil {
			return v, true
		}
	}
	return 0, false
}

func Int64(name string) (int64, bool) {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseInt(val, 0, 64); e == nil {
			return v, true
		}
	}
	return 0, false
}

func Bool(name string) (bool, bool) {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseBool(val); e == nil {
			return v, true
		}
	}
	return false, false
}
