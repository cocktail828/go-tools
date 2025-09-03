package httpx

import (
	"context"
	"net/http"
	"testing"

	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestHTTPX(t *testing.T) {
	_, err := Get(context.Background(), "https://baidu.com",
		Headers(map[string]string{"k": "v"}),
		Callback(func(r *http.Request) { assert.Equal(t, "v", r.Header.Get("k")) }),
	)
	z.Must(err)
}

func TestInsure(t *testing.T) {
	url := "https://aiportal.h3c.com:40212/snappyservice/profile/upload/ZJSZTB/virtualHuman.png"
	for _, b := range []bool{true, false} {
		InsecureSSL(b)
		_, err := http.DefaultClient.Get(url)
		assert.Nil(t, err)
	}
}
