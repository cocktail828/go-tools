package httpx

import (
	"bytes"
	"context"
	"net/http"
)

type option struct {
	body     []byte
	headers  map[string]string
	callback func(*http.Request)
}

type Option func(*option)

func apply(opts ...Option) *option {
	o := &option{headers: make(map[string]string)}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func Body(val []byte) Option {
	return func(o *option) { o.body = val }
}

func Headers(val map[string]string) Option {
	return func(o *option) {
		for k, v := range val {
			o.headers[k] = v
		}
	}
}

func Callback(cb func(*http.Request)) Option {
	return func(o *option) { o.callback = cb }
}

func Do(ctx context.Context, method string, url string, opts ...Option) (*Response, error) {
	o := apply(opts...)
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(o.body))
	if err != nil {
		return nil, err
	}

	for k, v := range o.headers {
		req.Header.Set(k, v)
	}

	if f := o.callback; f != nil {
		f(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return &Response{Response: resp}, nil
}

func Head(ctx context.Context, url string, opts ...Option) (*Response, error) {
	return Do(ctx, http.MethodHead, url, opts...)
}

func Get(ctx context.Context, url string, opts ...Option) (*Response, error) {
	return Do(ctx, http.MethodGet, url, opts...)
}

func Post(ctx context.Context, url string, opts ...Option) (*Response, error) {
	return Do(ctx, http.MethodPost, url, opts...)
}

func Put(ctx context.Context, url string, opts ...Option) (*Response, error) {
	return Do(ctx, http.MethodPut, url, opts...)
}

func Patch(ctx context.Context, url string, opts ...Option) (*Response, error) {
	return Do(ctx, http.MethodPatch, url, opts...)
}

func Delete(ctx context.Context, url string, opts ...Option) (*Response, error) {
	return Do(ctx, http.MethodDelete, url, opts...)
}
