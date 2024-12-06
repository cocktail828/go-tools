package queue_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/queue"
	"github.com/stretchr/testify/assert"
)

func TestQ(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Second*5, cancel)
	q := queue.WithContext(ctx)
	assert.Equal(t, 5, q.Concurrency())

	for i := 0; i < 10; i++ {
		q.Go(func() { fmt.Println(time.Now()); time.Sleep(time.Second / 2) })
	}
	q.Wait()
}
