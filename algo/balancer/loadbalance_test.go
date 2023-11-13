package balancer_test

import (
	"testing"

	"github.com/cocktail828/go-tools/algo/balancer"
	"github.com/stretchr/testify/assert"
)

type X struct {
	W int
	M int
}

type NNN []X

func (n NNN) Len() int {
	return len(n)
}

func (n NNN) Weight(idx int) int {
	return n[idx].W
}

func TestRR(t *testing.T) {
	lb := balancer.NewRoundRobin()
	assert.Equal(t, -1, lb.Get(NNN{}))
	assert.ElementsMatch(t, []int{1, 2, 3, 1, 2, 3}, func() []int {
		arr := NNN{X{5, 5}, X{3, 3}, X{2, 2}, X{1, 1}}
		res := []int{}
		for i := 0; i < 6; i++ {
			res = append(res, lb.Get(arr))
		}
		return res
	}())
}

func TestWRR(t *testing.T) {
	lb := balancer.NewWeightRoundRobin()
	assert.Equal(t, -1, lb.Get(NNN{}))
	assert.ElementsMatch(t, []int{3, 2, 1}, func() []int {
		arr := NNN{X{5, 5}, X{3, 3}, X{2, 2}, X{1, 1}}
		m := map[int]int{-1: 0, 5: 0, 3: 0, 2: 0, 1: 0}
		for i := 0; i < 6; i++ {
			m[lb.Get(arr)]++
		}
		return []int{m[3], m[2], m[1]}
	}())
}
