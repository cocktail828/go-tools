package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/cocktail828/go-tools/z/variadic"
)

type inVariadic struct{ variadic.Assigned }

type bodyKey struct{}

// populate request body
func Body(val []byte) variadic.Option { return variadic.SetValue(bodyKey{}, val) }
func (iv inVariadic) Body() []byte    { return variadic.GetValue[[]byte](iv, bodyKey{}) }

type headerKey struct{}

// populate HTTP headers
func Headers(val map[string]string) variadic.Option { return variadic.SetValue(headerKey{}, val) }
func (iv inVariadic) Headers() map[string]string {
	return variadic.GetValue[map[string]string](iv, headerKey{})
}

type CallbackFunc func(*http.Request)
type callbackKey struct{}

// user defined
func Callback(val CallbackFunc) variadic.Option { return variadic.SetValue(callbackKey{}, val) }
func (iv inVariadic) Callback() CallbackFunc {
	return variadic.GetValue[CallbackFunc](iv, callbackKey{})
}

func NewRequest(ctx context.Context, method string, url string, opts ...variadic.Option) (*http.Request, error) {
	iv := inVariadic{variadic.Compose(opts...)}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(iv.Body()))
	if err != nil {
		return nil, err
	}

	if f := iv.Callback(); f != nil {
		f(req)
	}

	return req, nil
}

type SimpleHTTP struct {
	Client    *http.Client
	Request   *http.Request
	Unmarshal func(status int, body []byte, i interface{}) error
}

func (c *SimpleHTTP) Fire(dst interface{}) error {
	if c.Unmarshal == nil {
		c.Unmarshal = func(_ int, body []byte, i interface{}) error {
			return json.Unmarshal(body, i)
		}
	}

	if c.Client == nil {
		c.Client = http.DefaultClient
	}

	resp, err := c.Client.Do(c.Request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return c.Unmarshal(resp.StatusCode, body, dst)
}
