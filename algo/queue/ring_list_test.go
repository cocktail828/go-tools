package queue

import (
	"fmt"
	"testing"
)

func TestRingList(t *testing.T) {
	rq := NewRingQueue(4)
	nodes := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, n := range nodes {
		rq.Insert(n)
	}

	for i := 0; i < 10; i++ {
		fmt.Println("===", rq.Poll())
	}

	for i := 0; i < 10; i++ {
		rq.Remove(rq.Poll())
	}

	for i := 0; i < 10; i++ {
		fmt.Println("###", rq.Poll())
	}
}
