package errcode

//go:generate stringer -type errCode -linecomment
type errCode int

const (
	GeneralErr errCode = -1 // unknow error
)

func (ec errCode) Code() int      { return int(ec) }
func (ec errCode) String() string { return "unknow error" }

func (ec errCode) ToError() *Error {
	return New(ec.Code(), ec.String())
}

func (ec errCode) WithError(err error) *Error {
	return New(ec.Code(), ec.String())
}

func (ec errCode) WithMessage(msg string) *Error {
	return New(ec.Code(), ec.String()).WithMessage(msg)
}

func (ec errCode) WithMessagef(format string, args ...interface{}) *Error {
	return New(ec.Code(), ec.String()).WithMessagef(format, args...)
}
