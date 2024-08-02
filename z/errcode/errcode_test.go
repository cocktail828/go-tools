package errcode_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z/errcode"
)

type errCode uint32

const (
	GeneralErr errCode = 10000 // unknow error
)

func (ec errCode) Code() uint32   { return uint32(ec) }
func (ec errCode) String() string { return "unknow error" }

func TestErrCode(t *testing.T) {
	fmt.Println(errcode.Errorf(GeneralErr, "asdfsdfg"))
	fmt.Println(errcode.Errorf(nil, "asdfsdfg"))
}
