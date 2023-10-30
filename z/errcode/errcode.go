package errcode

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type Error struct {
	mu   sync.Mutex
	err  error
	code int
}

func New() *Error {
	return &Error{}
}

func (e *Error) WithCode(c int) *Error {
	e.code = c
	return e
}

func (e *Error) WithError(err error) *Error {
	e.WithMessagef(err.Error())
	return e
}

func (e *Error) WithMessage(msg string) *Error {
	e.WithMessagef(msg)
	return e
}

func (e *Error) WithMessagef(format string, args ...string) *Error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.err == nil {
		e.err = errors.Errorf(format, args)
	} else {
		e.err = errors.WithMessagef(e.err, format, args)
	}
	return e
}

func (e *Error) Error() string {
	if e.IsNil() {
		return ""
	} else {
		return fmt.Sprintf("Code: %v, Msg: %v", e.code, e.err)
	}
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) IsNil() bool {
	return e.code == 0 && e.err == nil
}
