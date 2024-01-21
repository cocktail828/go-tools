package httpx_test

import (
	"context"
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/netx/httpx"
	"github.com/cocktail828/go-tools/z"
)

func TestHTTPX(t *testing.T) {
	c, err := httpx.NewWithContext(context.Background(), "https://baidu.com", httpx.WithMethod("GET"))
	z.Must(err)

	_, err = c.Fire()
	z.Must(err)

	var s string
	z.Must(c.ParseWith(httpx.Stringfy, &s))
	fmt.Println(s)
}

func TestGracefulServer(t *testing.T) {
	gs := &httpx.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go gs.ListenAndServe()
	fmt.Println(gs.WaitForSignal(time.Second, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT))
}
