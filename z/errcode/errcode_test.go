package errcode_test

import (
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/cocktail828/go-tools/z/errcode"
	"github.com/stretchr/testify/assert"
)

type errCode uint32

const (
	GeneralErr errCode = 10000 // unknow error
)

func (ec errCode) Code() uint32   { return uint32(ec) }
func (ec errCode) String() string { return "unknow error" }
func (ec errCode) WithMessagef(format string, args ...interface{}) *errcode.Error {
	return errcode.New(ec).WithMessagef(format, args...)
}

func (ec errCode) WithMessage(msg string) *errcode.Error {
	return errcode.New(ec).WithMessage(msg)
}

func (ec errCode) WithError(err error) *errcode.Error {
	return errcode.New(ec).WithError(err)
}

func TestErrCode(t *testing.T) {
	err := GeneralErr.WithMessage("asdfsdfg").WithError(io.ErrClosedPipe).WithError(net.ErrClosed)
	fmt.Println(err)
	assert.Equal(t, true, err.Is(io.ErrClosedPipe))
	assert.Equal(t, true, err.Is(net.ErrClosed))
}
