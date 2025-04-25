package main_test

import (
	"errors"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errcode uint32

//go:generate go run errorify.go -type errcode -linecomment
const (
	GeneralErr   errcode = iota // unknow error
	GeneralXrr1                 // unknow error
	GeneralXrr2                 // unknow error
	GeneralXrr3                 // unknow error
	GeneralXrr4                 // unknow error
	GeneralXrr5                 // unknow error
	GeneralXrr6                 // unknow error
	GeneralXrr7                 // unknow error
	GeneralXrr8                 // unknow error
	GeneralXrr9                 // unknow error
	GeneralXrr10                // unknow error
	GeneralXrr11                // unknow error
)

const (
	GeneralXrr20 errcode = 1000 // unknow error
	GeneralXrr21 errcode = 1002 // unknow error
)

func TestErrorify(t *testing.T) {
	err := GeneralXrr7.Wrap(io.ErrClosedPipe, "aklsjdf")
	assert.True(t, errors.Is(err, io.ErrClosedPipe))
	assert.False(t, errors.Is(err, net.ErrClosed))

	var e *Error
	if assert.True(t, errors.As(err, &e)) {
		assert.EqualValues(t, io.ErrClosedPipe, e.Cause())
		assert.EqualValues(t, GeneralXrr7.Code(), e.Code())
		assert.EqualValues(t, GeneralXrr7.Desc(), e.Desc())
	}
}
