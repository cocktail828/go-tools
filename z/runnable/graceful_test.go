package runnable

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestGS(t *testing.T) {
	gs := Graceful{
		Start: func() error {
			time.Sleep(time.Second * 1)
			return http.ErrBodyNotAllowed
		},
		Stop: func() error {
			return http.ErrContentLength
		},
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	fmt.Println(gs.Do(ctx))
}
