package errcode

import "fmt"

func Errorf(format string, args ...interface{}) error {
	return &errorString{
		s: fmt.Sprintf(format, args...),
	}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
