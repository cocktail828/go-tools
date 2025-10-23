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
