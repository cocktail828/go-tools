package reflectx

import (
	"fmt"
	"reflect"
	"strings"
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

func Stringify(obj any) string {
	visited := make(map[uintptr]struct{})
	return stringifyInternal(obj, 0, visited)
}

func stringifyInternal(obj any, indent int, visited map[uintptr]struct{}) string {
	if obj == nil {
		return "nil"
	}

	// prefer explicit interface implementation over reflect lookups
	if s, ok := obj.(interface{ Stringify() string }); ok {
		return s.Stringify()
	}

	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "nil"
		}

		ptrAddr := v.Pointer()
		if _, ok := visited[ptrAddr]; ok {
			return "<cycle detected>"
		}

		visited[ptrAddr] = struct{}{}

		elem := v.Elem()
		elemKind := elem.Kind()

		if elemKind != reflect.Struct {
			return stringifyInternal(elem.Interface(), indent, visited)
		}
		return "&" + stringifyInternal(elem.Interface(), indent, visited)
	}

	switch v.Kind() {
	case reflect.Struct:
		return stringifyStruct(v, indent, visited)
	case reflect.Map:
		return stringifyMap(v, indent, visited)
	case reflect.Slice, reflect.Array:
		return stringifySlice(v, indent, visited)
	case reflect.String:
		return "\"" + v.String() + "\""
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

func stringifyStruct(v reflect.Value, indent int, visited map[uintptr]struct{}) string {
	typ := v.Type()
	result := typ.Name() + "{\n"
	indent += 2

	hasFields := false
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := v.Field(i)

		// skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// check tag
		tag := field.Tag.Get("stringify")
		if tag == "false" {
			continue
		}

		hasFields = true
		result += strings.Repeat(" ", indent)
		result += field.Name + ": "

		if fieldVal.CanInterface() {
			result += stringifyInternal(fieldVal.Interface(), indent, visited)
		} else {
			result += "<unexported>"
		}
		result += ",\n"
	}

	indent -= 2
	if hasFields {
		result = result[:len(result)-2] + "\n"
	}
	result += strings.Repeat(" ", indent) + "}"
	return result
}

func stringifyMap(v reflect.Value, indent int, visited map[uintptr]struct{}) string {
	if v.IsNil() {
		return "nil"
	}

	result := "map[" + v.Type().Key().Kind().String() + "]" + v.Type().Elem().Kind().String() + "{\n"
	indent += 2

	iter := v.MapRange()
	first := true
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		if !first {
			result += ",\n"
		} else {
			first = false
		}

		result += strings.Repeat(" ", indent)
		if key.Kind() == reflect.String {
			result += "\"" + key.String() + "\": "
		} else {
			result += fmt.Sprintf("%v: ", key.Interface())
		}

		if val.CanInterface() {
			result += stringifyInternal(val.Interface(), indent, visited)
		} else {
			result += "<unexported>"
		}
	}

	indent -= 2
	result += "\n" + strings.Repeat(" ", indent) + "}"
	return result
}

func stringifySlice(v reflect.Value, indent int, visited map[uintptr]struct{}) string {
	if v.Kind() != reflect.Array && v.IsNil() {
		return "nil"
	}

	result := "[" + v.Type().Elem().Kind().String() + "]{"
	for i := 0; i < v.Len(); i++ {
		if elem := v.Index(i); elem.CanInterface() {
			result += stringifyInternal(elem.Interface(), indent, visited)
		} else {
			result += "<unexported>"
		}

		if i < v.Len()-1 {
			result += ", "
		}
	}

	result += "}"
	return result
}
