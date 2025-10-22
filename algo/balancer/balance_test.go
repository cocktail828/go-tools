package balancer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type X int

func (x X) Value() any    { return x }
func (x X) Weight() int   { return int(x) }
func (x X) Healthy() bool { return true }
func (x X) MarkFailure()  {}

func TestRR(t *testing.T) {
	assert.Equal(t, nil, NewRR(nil).Pick())
	lb := NewRR([]Node{X(3), X(2), X(1)})
	res := []any{}
	for range 6 {
		res = append(res, int(lb.Pick().Value().(X)))
	}

	assert.ElementsMatch(t, []any{3, 2, 1, 3, 2, 1}, res)
}

func TestWRR(t *testing.T) {
	assert.Equal(t, nil, NewWRR(nil).Pick())
	lb := NewWRR([]Node{X(3), X(2), X(1)})

	m := map[int]int{-1: 0, 3: 0, 2: 0, 1: 0}
	for range 6 {
		if v := lb.Pick(); v != nil {
			m[v.Weight()]++
		} else {
			m[-1]++
		}
	}

	assert.ElementsMatch(t, []any{3, 2, 1}, []any{m[3], m[2], m[1]})
}

func TestNodeArray(t *testing.T) {
	for n, b := range map[string]Balancer{
		"RR":       NewRR([]Node{X(3), X(2), X(1), X(4), X(5)}),
		"WRR":      NewWRR([]Node{X(3), X(2), X(1), X(4), X(5)}),
		"Failover": NewFailover([]Node{X(3), X(2), X(1), X(4), X(5)}),
		"Random":   NewRandom([]Node{X(3), X(2), X(1), X(4), X(5)}),
	} {
		t.Logf("%s: %+v", n, b)
		v := b.Pick()
		v.MarkFailure()
		t.Logf("%s: %+v %v", n, b, v)
	}
}
