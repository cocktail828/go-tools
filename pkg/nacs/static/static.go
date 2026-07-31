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

func (s *staticConfigor) Load(ctx context.Context) (nacs.ConfigInfo, error) {
	return nacs.ConfigInfo{
		DataID:  "static",
		Payload: s.payload,
	}, nil
}

func (s *staticConfigor) Monitor(cb func(nacs.ConfigInfo, error)) (context.CancelFunc, error) {
	return func() {}, nil
}

func (s *staticConfigor) Close() error {
	return nil
}
