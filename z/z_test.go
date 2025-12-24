package z

import (
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
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

func c() { fmt.Println(Stack(5)) }
func b() { c() }
func a() { b() }

func TestChainCall(t *testing.T) {
	ss := []string{}
	f := ChainCall(func(in int) error {
		ss = append(ss, "1")
		return nil
	}, func(in int) error {
		ss = append(ss, "2")
		return io.ErrClosedPipe
	}, func(in int) error {
		ss = append(ss, "3")
		return nil
	})

	f(0)
	assert.Equal(t, []string{"1", "2"}, ss)
}
