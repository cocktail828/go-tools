package z_test

import (
	"slices"
	"testing"

	"github.com/cocktail828/go-tools/z"
)

func TestZ(t *testing.T) {
	i := 0
	for {
		v := []int{1, 2, 3, 4, 5, 6}
		t.Log(slices.Delete(v, i, i+1))
		i++
		if i >= len(v) {
			break
		}
	}
	a()
}

func c() { z.DumpStack(5) }
func b() { c() }
func a() { b() }
