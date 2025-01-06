package stringx_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z/stringx"
)

func TestString(t *testing.T) {
	for i := 0; i < 20; i++ {
		fmt.Println(stringx.RandomName())
	}
}
