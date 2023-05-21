package werror

import (
	"sync"

	"github.com/pkg/errors"
)

type WrapperError struct {
	mu   sync.RWMutex
	errs []error
}

func New() *WrapperError {
	return &WrapperError{}
}

func (w *WrapperError) Add(err error) *WrapperError {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.errs = append(w.errs, err)
	return w
}

func (w *WrapperError) Error() error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if len(w.errs) == 0 {
		return nil
	}

	err := w.errs[0]
	for i := 1; i < len(w.errs); i++ {
		err = errors.WithMessage(err, w.errs[i].Error())
	}
	return err
}
