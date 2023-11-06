package errcode

import (
	"fmt"

	"github.com/pkg/errors"
)

type Error struct {
	code  int
	desc  string
	cause error
}

func New(code int, desc string) *Error {
	return &Error{code: code, desc: desc}
}

func (e *Error) WithError(err error) *Error {
	if e.cause == nil {
		e.cause = err
	} else {
		e.cause = errors.WithMessage(e.cause, err.Error())
	}
	return e
}

func (e *Error) WithMessage(msg string) *Error {
	if e.cause == nil {
		e.cause = errors.New(msg)
	} else {
		e.cause = errors.WithMessage(e.cause, msg)
	}
	return e
}

func (e *Error) WithMessagef(format string, args ...interface{}) *Error {
	if e.cause == nil {
		e.cause = errors.Errorf(format, args...)
	} else {
		e.cause = errors.WithMessagef(e.cause, format, args...)
	}
	return e
}

func (e *Error) Error() string {
	if e.cause == nil {
		return fmt.Sprintf("[Code:%v, Desc:'%v']", e.code, e.desc)
	}
	return fmt.Sprintf("[Code:%v, Msg:'%v']", e.code, e.desc) + ": " + e.cause.Error()
}

func (e *Error) Cause() error {
	return e.cause
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Desc() string {
	return e.desc
}
