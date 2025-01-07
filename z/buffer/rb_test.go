package buffer_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/cocktail828/go-tools/z/buffer"
)

func TestRB(t *testing.T) {
	rb := buffer.NewRingBuffer(10)

	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				item := fmt.Sprintf("producer-%d-item-%d", id, j)
				if err := rb.Enqueue(item); err != nil {
					fmt.Printf("Producer %d failed to enqueue: %s\n", id, err)
				} else {
					fmt.Printf("Producer %d enqueued: %s\n", id, item)
				}
			}
		}(i)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if item, err := rb.Dequeue(); err != nil {
					fmt.Printf("Consumer %d failed to dequeue: %s\n", id, err)
				} else {
					fmt.Printf("Consumer %d dequeued: %s\n", id, item)
				}
			}
		}(i)
	}
	wg.Wait()
}
