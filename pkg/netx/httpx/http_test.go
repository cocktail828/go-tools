package httpx_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cocktail828/go-tools/pkg/netx/httpx"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

var (
	rc httpx.RestClient
)

func TestHTTPX(t *testing.T) {
	_, err := rc.Get("https://baidu.com",
		httpx.Body([]byte("xxx")),
		httpx.Headers(map[string]string{"k": "v"}),
		httpx.Callback(func(r *http.Request) { fmt.Println("xxx", r.Header) }),
	)
	z.Must(err)
}

func TestInsure(t *testing.T) {
	url := "https://aiportal.h3c.com:40212/snappyservice/profile/upload/ZJSZTB/virtualHuman.png"
	httpx.InsecureSSL(false)
	_, err := http.DefaultClient.Get(url)
	assert.Nil(t, err)
	httpx.InsecureSSL(true)
	_, err = http.DefaultClient.Get(url)
	assert.Nil(t, err)
}
