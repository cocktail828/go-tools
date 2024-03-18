package environs

import (
	"os"
	"strconv"
)

func String(name string, dflt string) string {
	if val, ok := os.LookupEnv(name); ok {
		return val
	}
	return dflt
}

func Int64(name string, dflt int64) int64 {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseInt(val, 0, 64); e == nil {
			return v
		}
	}
	return dflt
}

func Bool(name string, dflt bool) bool {
	if val, ok := os.LookupEnv(name); ok {
		if v, e := strconv.ParseBool(val); e == nil {
			return v
		}
	}
	return dflt
}
