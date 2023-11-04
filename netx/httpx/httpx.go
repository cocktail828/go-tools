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

type Client struct {
	body    io.Reader
	headers map[string]string
	method  string
	request *http.Request
}

type Option func(*Client)

func WithHeaders(hs map[string]string) Option {
	return func(c *Client) {
		for k, v := range hs {
			c.headers[strings.ToLower(k)] = v
		}
	}
}

func WithBody(body []byte) Option {
	return func(c *Client) {
		c.body = bytes.NewBuffer(body)
	}
}

func WithMethod(m string) Option {
	return func(c *Client) {
		c.method = m
	}
}

func NewWithContext(ctx context.Context, url string, options ...Option) (*Client, error) {
	c := &Client{
		headers: map[string]string{"content-type": "application/json;charset=utf8"},
	}

	for _, f := range options {
		f(c)
	}

	req, err := http.NewRequestWithContext(ctx, c.method, url, c.body)
	if err != nil {
		return nil, err
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	c.request = req

	return c, nil
}

func (c *Client) Alter(f func(*http.Request)) {
	if f != nil {
		f(c.request)
	}
}

func (c *Client) Do(opts ...retry.Option) (resp *http.Response, err error) {
	retry.Do(func() error {
		resp, err = http.DefaultClient.Do(c.request)
		return err
	}, opts...)
	return
}

func (c *Client) ParseWith(parser Unmarshaler, resp *http.Response, i interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return parser(body, i)
}

func (c *Client) ParseBody(resp *http.Response, i interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.Unmarshal(body, i)
}

func Get(ctx context.Context, url string, options ...Option) (*http.Response, error) {
	req, err := NewWithContext(ctx, url, append(options, WithMethod("GET"))...)
	if err != nil {
		return nil, err
	}
	return req.Do()
}

func Put(ctx context.Context, url string, options ...Option) (*http.Response, error) {
	req, err := NewWithContext(ctx, url, append(options, WithMethod("PUT"))...)
	if err != nil {
		return nil, err
	}
	return req.Do()
}

func Post(ctx context.Context, url string, options ...Option) (*http.Response, error) {
	req, err := NewWithContext(ctx, url, append(options, WithMethod("POST"))...)
	if err != nil {
		return nil, err
	}
	return req.Do()
}

func Delete(ctx context.Context, url string, options ...Option) (*http.Response, error) {
	req, err := NewWithContext(ctx, url, append(options, WithMethod("DELETE"))...)
	if err != nil {
		return nil, err
	}
	return req.Do()
}
