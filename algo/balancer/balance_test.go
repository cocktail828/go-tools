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
	assert.Equal(t, nil, NewRoundRobin(nil).Pick())
	b := NewRoundRobin([]Node{X(3), X(2), X(1)})
	res := []any{}
	for range 6 {
		res = append(res, int(b.Pick().Value().(X)))
	}

	assert.ElementsMatch(t, []any{3, 2, 1, 3, 2, 1}, res)
}

func TestWRR(t *testing.T) {
	assert.Equal(t, nil, NewWeightRoundRobin(nil).Pick())
	b := NewWeightRoundRobin([]Node{X(3), X(2), X(1)})

	m := map[int]int{3: 0, 2: 0, 1: 0}
	for range 6000 {
		if v := b.Pick(); v != nil {
			m[v.Weight()]++
		}
	}
	assert.ElementsMatch(t, []any{3000, 2000, 1000}, []any{m[3], m[2], m[1]})
}

func TestCandidate(t *testing.T) {
	for _, b := range []Balancer{
		NewRoundRobin([]Node{X(3), X(2), X(1), X(4), X(5)}),
		NewWeightRoundRobin([]Node{X(3), X(2), X(1), X(4), X(5)}),
		NewFailover([]Node{X(3), X(2), X(1), X(4), X(5)}),
		NewRandom([]Node{X(3), X(2), X(1), X(4), X(5)}),
	} {
		old := make([]Node, len(b.Nodes()))
		copy(old, b.Nodes())
		v := b.Pick()
		v.MarkFailure()
		t.Logf("%-16s before: %v, after: %v, failed: %v", b, old, b.Nodes(), v)
	}
}
