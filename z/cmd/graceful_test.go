package cmd_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/cmd"
)

func TestGS(t *testing.T) {
	gs := cmd.Graceful{
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
