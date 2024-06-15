package httpx_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/netx/httpx"
	"github.com/cocktail828/go-tools/z"
)

type Mock struct{}

func (Mock) Register()   { log.Println("reg") }
func (Mock) DeRegister() { log.Println("dereg") }
func TestGracefulServer(t *testing.T) {
	gs := httpx.Server{
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

	_, err = http.DefaultClient.Do(req)
	z.Must(err)
}

func TestInsure(t *testing.T) {
	// httpx.SetInsure(true)
	url := "https://ddmedia-test.oss-cn-beijing.aliyuncs.com/ddmedia/test/mts/2024_06_05/3602013791191589/101/bfc939d1-2310-11ef-b243-06e43602a2c3.mp3"
	fmt.Println(http.DefaultClient.Get(url))
	os.Setenv("SSL_NO_VERIFY", "true")
	fmt.Println(http.DefaultClient.Get(url))
}
