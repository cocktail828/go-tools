package httpx

import (
	"bytes"
	"context"
	"net/http"

	"github.com/cocktail828/go-tools/z/variadic"
)

type bodyKey struct{}

// populate request body
func Body(val []byte) variadic.Option     { return variadic.Set(bodyKey{}, val) }
func getBody(c variadic.Container) []byte { return variadic.Value[[]byte](c, bodyKey{}) }

type headerKey struct{}

// populate HTTP headers
func Headers(val map[string]string) variadic.Option { return variadic.Set(headerKey{}, val) }
func getHeaders(c variadic.Container) map[string]string {
	return variadic.Value[map[string]string](c, headerKey{})
}

type CallbackFunc func(*http.Request)
type callbackKey struct{}

// user defined
func Callback(val CallbackFunc) variadic.Option { return variadic.Set(callbackKey{}, val) }
func getCallback(c variadic.Container) CallbackFunc {
	return variadic.Value[CallbackFunc](c, callbackKey{})
}

func Do(ctx context.Context, method string, url string, opts ...variadic.Option) (*Response, error) {
	iv := variadic.Compose(opts...)
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(getBody(iv)))
	if err != nil {
		return nil, err
	}

	for k, v := range getHeaders(iv) {
		req.Header.Set(k, v)
	}

	if f := getCallback(iv); f != nil {
		f(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return &Response{Response: resp}, nil
}

func Head(ctx context.Context, url string, opts ...variadic.Option) (*Response, error) {
	return Do(ctx, http.MethodHead, url, opts...)
}

func Get(ctx context.Context, url string, opts ...variadic.Option) (*Response, error) {
	return Do(ctx, http.MethodGet, url, opts...)
}

func Post(ctx context.Context, url string, opts ...variadic.Option) (*Response, error) {
	return Do(ctx, http.MethodPost, url, opts...)
}

func Put(ctx context.Context, url string, opts ...variadic.Option) (*Response, error) {
	return Do(ctx, http.MethodPut, url, opts...)
}

func Patch(ctx context.Context, url string, opts ...variadic.Option) (*Response, error) {
	return Do(ctx, http.MethodPatch, url, opts...)
}

func Delete(ctx context.Context, url string, opts ...variadic.Option) (*Response, error) {
	return Do(ctx, http.MethodDelete, url, opts...)
}
