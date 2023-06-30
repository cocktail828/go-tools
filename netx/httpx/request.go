package httpx

import (
	"context"
	"io"
	"net/http"
	"strings"
)

type httpOption struct {
	body    io.Reader
	headers map[string]string
	method  string
}

type option func(*httpOption)

func Headers(hs map[string]string) option {
	return func(ho *httpOption) {
		for k, v := range hs {
			ho.headers[strings.ToLower(k)] = v
		}
	}
}

func Body(body io.Reader) option {
	return func(ho *httpOption) { ho.body = body }
}

func Method(m string) option {
	return func(ho *httpOption) { ho.method = m }
}

func NewRequestWithContext(ctx context.Context, url string, options ...option) (*http.Request, error) {
	o := httpOption{
		headers: map[string]string{
			"content-type": "application/json;charset=utf8",
		},
	}

	for _, f := range options {
		f(&o)
	}

	req, err := http.NewRequestWithContext(ctx, o.method, url, o.body)
	if err != nil {
		return nil, err
	}

	for k, v := range o.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

func Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := NewRequestWithContext(ctx, url, Method("GET"))
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func Put(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	req, err := NewRequestWithContext(ctx, url, Method("PUT"), Body(body))
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func Post(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	req, err := NewRequestWithContext(ctx, url, Method("POST"), Body(body))
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func Delete(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	req, err := NewRequestWithContext(ctx, url, Method("DELETE"), Body(body))
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}
