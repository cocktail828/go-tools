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

func NewRequest(ctx context.Context, method string, url string, body io.Reader, opts ...Option) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for _, f := range opts {
		f(req)
	}

	return req, nil
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

type Unmarshal func([]byte, interface{}) error
type Parser struct {
	Unmarshal Unmarshal
}

func (p *Parser) Parse(reader io.ReadCloser, dst interface{}) error {
	if p.Unmarshal == nil {
		p.Unmarshal = json.Unmarshal
	}

	defer reader.Close()
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	return p.Unmarshal(body, dst)
}
