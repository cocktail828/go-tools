package httpx_test

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/netx/httpx"
	"github.com/cocktail828/go-tools/z"
)

func TestGracefulServer(t *testing.T) {
	gs := httpx.GracefulServer{
		Server: &http.Server{
			Addr:           ":8080",
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}
	log.Println(gs.ListenAndServe())
}

func TestHTTPX(t *testing.T) {
	req, err := httpx.NewRequestWithContext(context.Background(), "GET", "https://baidu.com", nil)
	z.Must(err)

	resp, err := http.DefaultClient.Do(req)
	z.Must(err)

	var s string
	z.Must(httpx.ParseWith(resp, httpx.Stringfy, &s))
	log.Println(s)
}
