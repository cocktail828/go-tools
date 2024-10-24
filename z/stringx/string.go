package stringx

import (
	"reflect"
	"unsafe"
)

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

// report whether 'array' contains string 's'
func Contains(array []string, s string) bool {
	for _, a := range array {
		if a == s {
			return true
		}
	}
	return false
}

// report whether 's' is a member of 'array'
func Oneof(s string, array ...string) bool {
	for _, a := range array {
		if a == s {
			return true
		}
	}
	return false
}
