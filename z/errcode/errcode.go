package errcode

import (
	"fmt"

	"github.com/pkg/errors"
)

// //go:generate stringer -type errCode -linecomment
// type errCode int
type ErrCode interface {
	Code() int
	String() string
}

type Error struct {
	ErrCode
	cause []error
}

func New(ec ErrCode) *Error {
	if ec == nil {
		return nil
	}
	return &Error{ErrCode: ec}
}

func (e *Error) WithError(err error) *Error {
	if e == nil {
		return nil
	}
	e.cause = append(e.cause, err)
	return e
}

func (e *Error) WithErrorf(format string, args ...interface{}) *Error {
	if e == nil {
		return nil
	}
	e.cause = append(e.cause, errors.Errorf(format, args...))
	return e
}

func (e *Error) WithMessage(msg string) *Error {
	if e == nil {
		return nil
	}
	e.cause = append(e.cause, errors.New(msg))
	return e
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("[%v#%v]@(%v)", e.Code(), e.String(), Join(e.cause...))
}

func (e *Error) Cause() error {
	if e == nil {
		return nil
	}
	return Join(e.cause...)
}
