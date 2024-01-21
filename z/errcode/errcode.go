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
	return &Error{
		ecode: ec,
		err:   errors.New(ec.String()),
	}
}

func (e *Error) WithMessage(msg string) *Error {
	e.err = errors.WithMessage(e.err, msg)
	return e
}

func (e *Error) WithMessagef(format string, args ...interface{}) *Error {
	e.err = errors.WithMessagef(e.err, format, args...)
	return e
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v@%v", e.ecode.Code(), e.err)
}

func (e *Error) Cause() error {
	return errors.Cause(e.err)
}

func (e *Error) Code() int {
	return e.ecode.Code()
}

func (e *Error) Desc() string {
	return e.ecode.String()
}
