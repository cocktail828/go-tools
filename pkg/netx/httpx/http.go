package httpx

import (
	"bytes"
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

type RestClient struct {
	http.Client
}

func (rc RestClient) Do(method string, url string, opts ...variadic.Option) (*http.Response, error) {
	iv := variadic.Compose(opts...)
	req, err := http.NewRequest(method, url, bytes.NewReader(getBody(iv)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range getHeaders(iv) {
		req.Header.Set(k, v)
	}

	if f := getCallback(iv); f != nil {
		f(req)
	}

	return rc.Client.Do(req)
}

func (rc RestClient) Head(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodHead, url, opts...)
}

func (rc RestClient) Get(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodGet, url, opts...)
}

func (rc RestClient) Post(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodPost, url, opts...)
}

func (rc RestClient) PostForm(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodPost, url, append(opts,
		Headers(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		}),
	)...)
}

func (rc RestClient) Put(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodPut, url, opts...)
}

func (rc RestClient) Patch(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodPatch, url, opts...)
}

func (rc RestClient) Delete(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodDelete, url, opts...)
}
