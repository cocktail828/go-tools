package balancer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type X int

func (x X) Value() any  { return x }
func (x X) Weight() int { return int(x) }

func TestRR(t *testing.T) {
	assert.Equal(t, nil, NewRR(nil).Pick())
	lb := NewRR([]Node{X(3), X(2), X(1)})
	assert.ElementsMatch(t, []any{X(3), X(2), X(1), X(3), X(2), X(1)}, func() []any {
		res := []any{}
		for i := 0; i < 6; i++ {
			res = append(res, lb.Pick())
		}
		return res
	}())
}

func TestWRR(t *testing.T) {
	assert.Equal(t, nil, NewWRR(nil).Pick())
	lb := NewWRR([]Node{X(3), X(2), X(1)})
	assert.ElementsMatch(t, []any{3, 2, 1}, func() []any {
		m := map[int]int{-1: 0, 3: 0, 2: 0, 1: 0}
		for i := 0; i < 6; i++ {
			if v := lb.Pick(); v != nil {
				m[v.Weight()]++
			} else {
				m[-1]++
			}
		}
		return []any{m[3], m[2], m[1]}
	}())
}
