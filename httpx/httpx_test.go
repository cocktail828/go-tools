package httpx_test

import (
	"context"
	"testing"

	"github.com/cocktail828/go-tools/httpx"
)

func TestHTTPX(t *testing.T) {
	resp, err := httpx.Get(context.Background(), "https://baidu.com")
	if err != nil {
		panic(err)
	}

	var s string
	if err = httpx.NewResponseParser(httpx.Stringfy).Parse(resp, &s); err != nil {
		panic(err)
	}
}
