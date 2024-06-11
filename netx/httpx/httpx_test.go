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
	srv := &httpx.GoHTTPServer{
		Server: http.Server{
			Addr:           ":0",
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

	gs := httpx.GracefulServer{
		Server:  srv,
		Signals: httpx.DefaultSignals,
		Timeout: time.Second * 3,
	}
	go func() {
		<-time.After(time.Second)
		log.Println("port", srv.Port())
	}()
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
