package caller

import (
	"runtime"
)

func Current() *runtime.Func {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	return runtime.FuncForPC(pc[0])
}
