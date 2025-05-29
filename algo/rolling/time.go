package rolling

import _ "unsafe"

var (
	// for debug
	unixNano = func() int64 { return runtimeNano() }
)

// runtimeNano returns the current value of the runtime clock in nanoseconds.
//
//go:linkname runtimeNano runtime.nanotime
func runtimeNano() int64

func SetTime(f func() int64)    { unixNano = f }
func round(tm, gap int64) int64 { return (tm / gap) * gap }
