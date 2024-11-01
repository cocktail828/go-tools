package z_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z"
)

func TestZ(t *testing.T) {
	v := []int{1, 2, 3, 4, 5, 6}
	for i := -3; i < len(v)+3; i++ {
		fmt.Println(z.Delete(v, i))
	}
}
