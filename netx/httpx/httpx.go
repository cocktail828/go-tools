package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/avast/retry-go/v4"
)

type Unmarshaler func([]byte, interface{}) error

func Stringfy(b []byte, i interface{}) error {
	if s, ok := i.(*string); ok {
		*s = string(b)
		return nil
	}

	return fmt.Errorf("type assert fail: %T not *string", i)
}

type simpleHTTP struct {
	body    io.Reader
	headers map[string]string
	method  string
	request *http.Request
}

type option func(*simpleHTTP)

func Headers(hs map[string]string) option {
	return func(sh *simpleHTTP) {
		for k, v := range hs {
			sh.headers[strings.ToLower(k)] = v
		}
	}
}

func Body(body []byte) option {
	return func(sh *simpleHTTP) {
		sh.body = bytes.NewBuffer(body)
	}
}

func Method(m string) option {
	return func(sh *simpleHTTP) {
		sh.method = m
	}
}

func NewWithContext(ctx context.Context, url string, options ...option) (*simpleHTTP, error) {
	sh := &simpleHTTP{
		headers: map[string]string{"content-type": "application/json;charset=utf8"},
	}

	for _, f := range options {
		f(sh)
	}

	req, err := http.NewRequestWithContext(ctx, sh.method, url, sh.body)
	if err != nil {
		return nil, err
	}

	for k, v := range sh.headers {
		req.Header.Set(k, v)
	}
	sh.request = req

	return sh, nil
}

func (sh *simpleHTTP) Do(opts ...retry.Option) (resp *http.Response, err error) {
	retry.Do(func() error {
		resp, err = http.DefaultClient.Do(sh.request)
		return err
	}, append([]retry.Option{retry.Attempts(3)}, opts...)...)
	return
}

func (sh *simpleHTTP) ParseWith(parser Unmarshaler, resp *http.Response, i interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return parser(body, i)
}

func (sh *simpleHTTP) ParseBody(resp *http.Response, i interface{}) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.Unmarshal(data, i)
}

func Get(ctx context.Context, url string, options ...option) (*http.Response, error) {
	req, err := NewWithContext(ctx, url, append(options, Method("GET"))...)
	if err != nil {
		return nil, err
	}
	return req.Do()
}

func Put(ctx context.Context, url string, options ...option) (*http.Response, error) {
	req, err := NewWithContext(ctx, url, append(options, Method("PUT"))...)
	if err != nil {
		return nil, err
	}
	return req.Do()
}

func Post(ctx context.Context, url string, options ...option) (*http.Response, error) {
	req, err := NewWithContext(ctx, url, append(options, Method("POST"))...)
	if err != nil {
		return nil, err
	}
	return req.Do()
}

func Delete(ctx context.Context, url string, options ...option) (*http.Response, error) {
	req, err := NewWithContext(ctx, url, append(options, Method("DELETE"))...)
	if err != nil {
		return nil, err
	}
	return req.Do()
}
