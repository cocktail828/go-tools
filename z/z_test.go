package z_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/cocktail828/go-tools/z"
)

func TestZ(t *testing.T) {
	i := 0
	for {
		v := []int{1, 2, 3, 4, 5, 6}
		fmt.Println(slices.Delete(v, i, i+1))
		i++
		if i >= len(v) {
			break
		}
	}
	a()
}

func c() { z.DumpStack(5, 1) }
func b() { c() }
func a() { b() }
