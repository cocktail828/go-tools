package diagnostic

import (
	"fmt"

	"github.com/pkg/errors"
)

type diagnostic []error

func New() diagnostic {
	return diagnostic{}
}

func (d diagnostic) WithError(err error) diagnostic {
	return append(d, err)
}

func (d diagnostic) WithMessage(msg string) diagnostic {
	return append(d, errors.New(msg))
}

func (d diagnostic) WithMessagef(format string, args ...interface{}) diagnostic {
	return append(d, fmt.Errorf(format, args...))
}

func (d diagnostic) HasError() bool {
	return len(d) > 0
}

func (d diagnostic) As(target interface{}) bool {
	for i := 0; i < len(d); i++ {
		if errors.As(d[i], target) {
			return true
		}
	}
	return false
}

func (d diagnostic) Is(target error) bool {
	for i := 0; i < len(d); i++ {
		if errors.Is(d[i], target) {
			return true
		}
	}
	return false
}

func (d diagnostic) Error() string {
	count := len(d)
	switch {
	case count == 0:
		return ""
	case count == 1:
		return d[0].Error()
	default:
		return fmt.Sprintf("%s, and %d other diagnostic(s)", d[0].Error(), count-1)
	}
}

func (d diagnostic) ToError() error {
	count := len(d)
	switch {
	case count == 0:
		return nil
	case count == 1:
		return d[0]
	default:
		return fmt.Errorf("%v, and %d other diagnostic(s)", d[0], count-1)
	}
}
