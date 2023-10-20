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
	err := d.ToError()
	if err == nil {
		return ""
	}
	return err.Error()
}

func (d diagnostic) ToError() error {
	switch len(d) {
	case 0:
		return nil
	default:
		err := d[0]
		for i := 1; i < len(d); i++ {
			err = errors.WithMessage(err, d[i].Error())
		}
		return err
	}
}
