package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingList(t *testing.T) {
	cnt := 4
	rq := NewRingQueue(cnt)
	nodes := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i, n := range nodes {
		if i < cnt {
			assert.True(t, rq.Insert(n))
		} else {
			assert.False(t, rq.Insert(n))
		}
	}

	for i := 0; i < 10; i++ {
		assert.Equal(t, i%4+1, rq.Poll().Value.(int))
	}

	for i := 0; i < 10; i++ {
		rq.Remove(rq.Poll())
	}

	for i := 0; i < 10; i++ {
		assert.Nil(t, rq.Poll())
	}
}
