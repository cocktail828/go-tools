package reflectx

import (
	"reflect"
	"unsafe"
)

func IsNil(obj interface{}) bool {
	if obj == nil {
		return true
	}

	return reflect.ValueOf(obj).IsNil()
}

func BytesToString(b []byte) string {
	byteHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	strHeader := reflect.StringHeader{
		Data: byteHeader.Data,
		Len:  byteHeader.Len,
	}

	return *(*string)(unsafe.Pointer(&strHeader))
}

func StringToBytes(s string) []byte {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	byteHeader := reflect.SliceHeader{
		Data: strHeader.Data,
		Len:  strHeader.Len,
		Cap:  strHeader.Len,
	}

	return *(*[]byte)(unsafe.Pointer(&byteHeader))
}
