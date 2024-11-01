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

func Oneof(s string, set []string) bool {
	for _, a := range set {
		if a == s {
			return true
		}
	}
	return false
}

func Unique(s []string) []string {
	m := map[string]struct{}{}
	r := []string{}
	for _, k := range s {
		if _, has := m[k]; !has {
			m[k] = struct{}{}
			r = append(r, k)
		}
	}
	return r
}

func Subset(test, base []string) bool {
	for _, k := range test {
		if !Oneof(k, base) {
			return false
		}
	}
	return true
}

func EqualValues(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for _, k := range s1 {
		if !Oneof(k, s2) {
			return false
		}
	}
	return true
}
