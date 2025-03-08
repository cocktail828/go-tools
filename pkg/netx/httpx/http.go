package httpx

import (
	"bytes"
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

func addHeaders(header map[string]string, opts []variadic.Option) []variadic.Option {
	iv := inVariadic{variadic.Compose(opts...)}
	for k, v := range iv.Headers() {
		header[k] = v
	}
	return append(opts, Headers(header))
}

type RestClient struct {
	http.Client
}

func (rc RestClient) Do(method string, url string, opts ...variadic.Option) (*http.Response, error) {
	iv := inVariadic{variadic.Compose(opts...)}
	req, err := http.NewRequest(method, url, bytes.NewReader(iv.Body()))
	if err != nil {
		return nil, err
	}

	for k, v := range iv.Headers() {
		req.Header.Add(k, v)
	}

	if f := iv.Callback(); f != nil {
		f(req)
	}

	return rc.Client.Do(req)
}

func (rc RestClient) Head(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodHead, url, addHeaders(map[string]string{
		"Content-Type": "application/json",
	}, opts)...)
}

func (rc RestClient) Get(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodGet, url, addHeaders(map[string]string{
		"Content-Type": "application/json",
	}, opts)...)
}

func (rc RestClient) Post(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodPost, url, addHeaders(map[string]string{
		"Content-Type": "application/json",
	}, opts)...)
}

func (rc RestClient) PostForm(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodPost, url, addHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, opts)...)
}

func (rc RestClient) Put(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodPut, url, addHeaders(map[string]string{
		"Content-Type": "application/json",
	}, opts)...)
}

func (rc RestClient) Patch(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodPatch, url, addHeaders(map[string]string{
		"Content-Type": "application/json",
	}, opts)...)
}

func (rc RestClient) Delete(url string, opts ...variadic.Option) (*http.Response, error) {
	return rc.Do(http.MethodDelete, url, addHeaders(map[string]string{
		"Content-Type": "application/json",
	}, opts)...)
}
