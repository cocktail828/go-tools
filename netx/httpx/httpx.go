package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	body     io.Reader
	headers  map[string]string
	method   string
	request  *http.Request
	response *http.Response
}

type Option func(*Client)

func WithHeaders(hs map[string]string) Option {
	return func(c *Client) {
		for k, v := range hs {
			c.headers[k] = v
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
	f(c.request)
}

func (c *Client) Fire() (*http.Response, error) {
	var err error
	c.response, err = http.DefaultClient.Do(c.request)
	return c.response, err
}

func (c *Client) ParseWith(parser Unmarshaler, dst interface{}) error {
	body, err := io.ReadAll(c.response.Body)
	if err != nil {
		return err
	}
	defer c.response.Body.Close()
	return parser(body, dst)
}

func (c *Client) Parse(dst interface{}) error {
	body, err := io.ReadAll(c.response.Body)
	if err != nil {
		return err
	}
	defer c.response.Body.Close()
	return json.Unmarshal(body, dst)
}
