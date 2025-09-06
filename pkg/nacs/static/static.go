package static

import (
	"context"

	"github.com/cocktail828/go-tools/pkg/nacs"
)

type staticConfigor struct {
	payload []byte
}

func NewStaticConfigor(payload []byte) nacs.Configor {
	return &staticConfigor{
		payload: payload,
	}
}

func (s *staticConfigor) Get(opts ...nacs.GetOpt) ([]byte, error) {
	return s.payload, nil
}

func (s *staticConfigor) Monitor(cb nacs.OnChange, opts ...nacs.MonitorOpt) (context.CancelFunc, error) {
	return func() {}, nil
}

func (s *staticConfigor) Close() error {
	return nil
}
