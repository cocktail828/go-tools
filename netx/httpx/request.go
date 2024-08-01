package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/cocktail828/go-tools/z/locker"
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

var (
	unmarshallersMu = sync.RWMutex{}
	unmarshallers   = map[string]Unmarshal{}
)

func RegisterUnmarshal(name string, f Unmarshal) {
	if f == nil {
		return
	}
	locker.WithLock(&unmarshallersMu, func() { unmarshallers[name] = f })
}

type Unmarshal func([]byte, interface{}) error
type Parser struct {
	Unmarshal Unmarshal
}

func (p *Parser) Parse(reader io.ReadCloser, dst interface{}) error {
	if p.Unmarshal == nil {
		locker.WithRLock(&unmarshallersMu, func() {
			if v, ok := unmarshallers["json"]; ok {
				p.Unmarshal = v
			}
		})

		if p.Unmarshal == nil {
			return errors.New("no unmarshaller")
		}
	}

	defer reader.Close()
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	return p.Unmarshal(body, dst)
}
