package errcode_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z/errcode"
)

func TestXxx(t *testing.T) {
	e := errcode.New()
	e.WithMessage("ajsdjf").WithMessage("qeqwe")
	fmt.Println(e)
}
