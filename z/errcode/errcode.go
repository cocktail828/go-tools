package errcode

import (
	"fmt"
	"strings"
)

type Error struct {
	code   int
	desc   string
	errMsg []string
}

func New() *Error {
	return &Error{}
}

func (e *Error) WithCode(c int) *Error {
	e.code = c
	return e
}

func (e *Error) WithDesc(s string) *Error {
	e.desc = s
	return e
}

func (e *Error) WithError(err error) *Error {
	e.errMsg = append(e.errMsg, err.Error())
	return e
}

func (e *Error) WithMessage(msg string) *Error {
	e.errMsg = append(e.errMsg, msg)
	return e
}

func (e *Error) WithMessagef(format string, args ...string) *Error {
	e.errMsg = append(e.errMsg, fmt.Sprintf(format, args))
	return e
}

func (e *Error) Error() string {
	if e.IsNil() {
		return ""
	}

	errmsg := []string{}
	if e.desc == "" {
		errmsg = append(errmsg, fmt.Sprintf("code: %v", e.code))
	} else {
		errmsg = append(errmsg, fmt.Sprintf("code: %v, desc: %v", e.code, e.desc))
	}
	for pos, msg := range e.errMsg {
		errmsg = append(errmsg, fmt.Sprintf("error#%v: %v", pos, msg))
	}
	return strings.Join(errmsg, "\n")
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) IsNil() bool {
	return e.code == 0 && len(e.errMsg) == 0
}
