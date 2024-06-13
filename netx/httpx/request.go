package httpx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type Option func(*http.Request)

func WithHeaders(hs map[string]string) Option {
	return func(r *http.Request) {
		for k, v := range hs {
			r.Header.Set(k, v)
		}
	}
}

func NewRequestWithContext(ctx context.Context, method string, url string, body io.Reader, opts ...Option) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	for _, f := range opts {
		f(req)
	}

	return req, nil
}

func ParseWith(resp *http.Response, parser func([]byte, interface{}) error, dst interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return parser(body, dst)
}

func Parse(resp *http.Response, dst interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}
