package httpx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/netx/httpx"
)

func TestHTTPX(t *testing.T) {
	sh, err := httpx.NewWithContext(context.Background(), "https://baidu.com", httpx.Method("GET"))
	if err != nil {
		panic(err)
	}

	resp, err := sh.Do()
	if err != nil {
		panic(err)
	}

	var s string
	if err := sh.ParseWith(httpx.Stringfy, resp, &s); err != nil {
		panic(err)
	}
	fmt.Println(s)
}
