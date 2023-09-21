package z

import "unsafe"

type interfaceStructure struct {
	pt uintptr
	pv uintptr
}

func IsNil(obj interface{}) bool {
	if obj == nil {
		return true
	}
	return (*interfaceStructure)(unsafe.Pointer(&obj)).pv == 0
}
