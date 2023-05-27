package httpx_test

import (
	"context"
	"io"
	"testing"

	"github.com/cocktail828/go-tools/httpx"
)

func TestHTTPX(t *testing.T) {
	resp, err := httpx.Get(context.Background(), "https://baidu.com")
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	_, err = httpx.ParseBody[string](body, httpx.Stringfy)
	if err != nil {
		panic(err)
	}
}
