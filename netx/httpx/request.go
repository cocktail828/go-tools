package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// define a real type of Option, to make sure the type is exactly what we want.
type Option struct{ e any }

func WithBody(d []byte) Option {
	return Option{bytes.NewReader(d)}
}

func WithHeaders(headers map[string]string) Option {
	return Option{func(r *http.Request) {
		for k, v := range headers {
			r.Header.Set(k, v)
		}
	}}
}

func WithModifier(f func(*http.Request)) Option {
	return Option{f}
}

func NewRequest(ctx context.Context, method string, url string, opts ...Option) (*http.Request, error) {
	var reader io.Reader
	for _, o := range opts {
		if o.e != nil {
			if val, ok := o.e.(io.Reader); ok {
				reader = val
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}

	for _, o := range opts {
		if o.e != nil {
			if val, ok := o.e.(func(*http.Request)); ok {
				val(req)
			}
		}
	}

	return req, nil
}

type Caller struct {
	Client    *http.Client
	Request   *http.Request
	Unmarshal func([]byte, interface{}) error
}

func (c *Caller) Do(dst interface{}) error {
	if c.Unmarshal == nil {
		c.Unmarshal = json.Unmarshal
	}

	if c.Client == nil {
		c.Client = http.DefaultClient
	}

	resp, err := c.Client.Do(c.Request)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return c.Unmarshal(body, dst)
}
