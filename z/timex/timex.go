package timex

import (
	_ "unsafe"
)

var (
	UnixNano = func() int64 { return runtimeNano() } // for debug
)

// runtimeNano returns the current value of the runtime clock in nanoseconds.
//
//go:linkname runtimeNano runtime.nanotime
func runtimeNano() int64

func ResetTime() { UnixNano = runtimeNano }

func SetTime(f func() int64) { UnixNano = f }
