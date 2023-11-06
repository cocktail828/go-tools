package errcode

import (
	"fmt"

	"github.com/pkg/errors"
)

type Error struct {
	ecode errCode
	cause error
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
		return ""
	}
	return fmt.Sprintf("[Code:%d, Msg:'%v']", int(e.ecode), e.ecode.String()) + ": " + e.cause.Error()
}

func (e *Error) Cause() error {
	return e.cause
}

func (e *Error) Code() int {
	return int(e.ecode)
}

func (e *Error) Desc() string {
	return e.ecode.String()
}
