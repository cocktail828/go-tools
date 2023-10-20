package compare

import (
	"bytes"
	"reflect"
)

func isFunction(arg interface{}) bool {
	if arg == nil {
		return false
	}
	return reflect.TypeOf(arg).Kind() == reflect.Func
}

func equalObjects(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}
	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}

// two objects are equal.
//
//	assert.Equal(t, 123, 123)
//
// Pointer variable equality is determined based on the equality of the
// referenced values (as opposed to the memory addresses). Function equality
// cannot be determined and will always fail.
func Equal(expected, actual interface{}) bool {
	if expected == nil && actual == nil {
		return true
	}
	// cannot take func type as argument
	if isFunction(expected) || isFunction(actual) {
		return false
	}
	return equalObjects(expected, actual)
}

// EqualValues gets whether two objects are equal, or if their values are equal.
func EqualValues(expected, actual interface{}) bool {
	if equalObjects(expected, actual) {
		return true
	}
	actualType := reflect.TypeOf(actual)
	if actualType == nil {
		return false
	}
	expectedValue := reflect.ValueOf(expected)
	// Attempt comparison after type conversion
	return expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) &&
		reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
}
