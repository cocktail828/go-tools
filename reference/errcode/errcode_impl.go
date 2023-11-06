package errcode

//go:generate stringer -type errCode -linecomment
type errCode int

func (ec errCode) Code() int { return int(ec) }

func (errCode) String() string { return "" }

const (
	GeneralErr errCode = -1 // unknow error
)

func (ec errCode) WithError(err error) *Error {
	return &Error{ecode: ec, cause: err}
}

func (ec errCode) WithMessage(msg string) *Error {
	e := &Error{ecode: ec}
	return e.WithMessage(msg)
}

func (ec errCode) WithMessagef(format string, args ...interface{}) *Error {
	e := &Error{ecode: ec}
	return e.WithMessagef(format, args...)
}
