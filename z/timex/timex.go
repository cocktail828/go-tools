package timex

import (
	_ "unsafe"
)

var (
	nanotime = func() int64 { return runtimeNano() } // for debug
)

// runtimeNano returns the current value of the runtime clock in nanoseconds.
//
//go:linkname runtimeNano runtime.nanotime
func runtimeNano() int64

func ResetTime() { nanotime = runtimeNano }

func SetTime(f func() int64) { nanotime = f }

func UnixNano() int64 { return nanotime() }
