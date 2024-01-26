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
	ecode ErrCode
	err   error
}

func New(ec ErrCode) *Error {
	if ec == nil {
		return nil
	}
	return &Error{
		ecode: ec,
		err:   errors.New(ec.String()),
	}
}

func (e *Error) WithMessage(msg string) *Error {
	if e == nil {
		return nil
	}
	e.err = errors.WithMessage(e.err, msg)
	return e
}

func (e *Error) WithMessagef(format string, args ...interface{}) *Error {
	if e == nil {
		return nil
	}
	e.err = errors.WithMessagef(e.err, format, args...)
	return e
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%v@%v", e.ecode.Code(), e.err)
}

func (e *Error) Cause() error {
	if e == nil {
		return nil
	}
	return errors.Cause(e.err)
}

func (e *Error) Code() int {
	if e == nil {
		return 0
	}
	return e.ecode.Code()
}

func (e *Error) Desc() string {
	if e == nil {
		return ""
	}
	return e.ecode.String()
}
