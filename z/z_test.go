package z

import (
	"slices"
	"testing"
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

func c() { DumpStack(5) }
func b() { c() }
func a() { b() }
