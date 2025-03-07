package errorx

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type Code interface {
	Code() uint32
	Desc() string
}

type Error struct {
	code  Code
	cause error
}

func New(c Code, err error) error {
	return &Error{code: c, cause: err}
}

func (w Error) Error() string {
	if w.cause == nil || w.code == nil {
		return ""
	}
	return fmt.Sprintf("code: %d, desc: %q, cause: \"%v\"", w.code.Code(), w.code.Desc(), w.cause)
}

func (w *Error) Code() uint32 { return w.code.Code() }
func (w *Error) Desc() string { return w.code.Desc() }
func (w *Error) Cause() error { return errors.Cause(w.cause) }

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *Error) Unwrap() error { return w.cause }

func (w *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\ncode: %d desc: %s", w.Cause(), w.Code(), w.Desc())
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}
