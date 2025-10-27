package reflectx

import (
	"reflect"
	"unsafe"
)

func IsNil(obj any) bool {
	if obj == nil {
		return true
	}

	v := reflect.ValueOf(obj)
	kind := v.Kind()
	switch kind {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

func BytesToString(b []byte) string {
	// NOTE: The returned string is only valid until the next call to BytesToString.
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

func StringToBytes(s string) []byte {
	// NOTE: The returned slice is only valid until the next call to StringToBytes or BytesToString.
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
