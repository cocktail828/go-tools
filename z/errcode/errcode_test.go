package errcode_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z/errcode"
)

//go:generate stringer -type errCode -linecomment
type errCode int

const (
	GeneralErr errCode = -1 // unknow error
)

func (ec errCode) Code() int      { return int(ec) }
func (ec errCode) String() string { return "unknow error" }

func TestErrCode(t *testing.T) {
	e := errcode.New(GeneralErr)
	// fmt.Println(e.Code(), e.String(), e.Cause())

	e.WithMessage("xxx").WithMessage("xajsdf")
	// fmt.Println(e.Code(), e.String(), e.Cause())
	fmt.Println(e)
}
