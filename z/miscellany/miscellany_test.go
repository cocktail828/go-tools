package miscellany_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z/miscellany"
)

func TestString(t *testing.T) {
	for i := 0; i < 20; i++ {
		fmt.Println(miscellany.RandomName(miscellany.WithCase(), miscellany.WithWidth(20)))
	}
}
