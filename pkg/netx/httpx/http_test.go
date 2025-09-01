package httpx

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

var (
	rc RestClient
)

func TestHTTPX(t *testing.T) {
	_, err := rc.Get("https://baidu.com",
		Body([]byte("xxx")),
		Headers(map[string]string{"k": "v"}),
		Callback(func(r *http.Request) { fmt.Println("xxx", r.Header) }),
	)
	z.Must(err)
}

func TestInsure(t *testing.T) {
	url := "https://aiportal.h3c.com:40212/snappyservice/profile/upload/ZJSZTB/virtualHuman.png"
	InsecureSSL(false)
	_, err := http.DefaultClient.Get(url)
	assert.Nil(t, err)
	InsecureSSL(true)
	_, err = http.DefaultClient.Get(url)
	assert.Nil(t, err)
}
