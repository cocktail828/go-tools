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

type Mock struct{}

func (Mock) Register()   { log.Println("reg") }
func (Mock) DeRegister() { log.Println("dereg") }
func TestGracefulServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	gs := httpx.Server{
		Context: ctx,
		Cancel:  cancel,
		Server: http.Server{
			Addr:           ":8080",
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}
	go func() {
		log.Println(gs.ListenAndServe())
	}()

	<-time.After(time.Second)
	log.Println("port", gs.Port())
	gs.WaitForSignal(Mock{}, time.Second, httpx.DefaultSignals...)
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
