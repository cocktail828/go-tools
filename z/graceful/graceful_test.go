package graceful_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/graceful"
)

func TestGS(t *testing.T) {
	gs := graceful.Graceful{
		Start: func() error {
			time.Sleep(time.Second * 1)
			return http.ErrBodyNotAllowed
		},
		Stop: func() error {
			return http.ErrContentLength
		},
	}
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*3)
	fmt.Println(gs.Do(ctx))
}
