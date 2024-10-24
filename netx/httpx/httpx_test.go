package httpx_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/cocktail828/go-tools/netx/httpx"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestHTTPX(t *testing.T) {
	req, err := httpx.NewRequest(context.Background(), "GET", "https://baidu.com",
		httpx.WithBody([]byte("xxx")),
		httpx.WithHeaders(map[string]string{"k": "v"}),
		httpx.WithModifier(func(r *http.Request) { fmt.Println("xxx") }),
	)
	z.Must(err)

	_, err = http.DefaultClient.Do(req)
	z.Must(err)
}

func TestInsure(t *testing.T) {
	url := "https://aiportal.h3c.com:40212/snappyservice/profile/upload/ZJSZTB/virtualHuman.png"
	httpx.InsecureSSL(false)
	_, err := http.DefaultClient.Get(url)
	httpx.InsecureSSL(true)
	_, err = http.DefaultClient.Get(url)
	assert.Nil(t, err)
}
