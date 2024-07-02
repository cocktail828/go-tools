package httpx_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/cocktail828/go-tools/netx/httpx"
	"github.com/cocktail828/go-tools/z"
)

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
