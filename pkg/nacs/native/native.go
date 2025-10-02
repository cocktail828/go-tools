package native

import (
	"context"
	"net/url"
	"os"
	"sync"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/pkg/errors"
	"gopkg.in/fsnotify.v1"
)

type fileConfigor struct {
	rctx    context.Context
	rcancel context.CancelFunc

	mu      sync.RWMutex
	fpath   string // 文件路径
	payload []byte // 文件内容
}

// native://localhost?/path1
func NewNativeConfigor(u *url.URL) (nacs.Configor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	f := &fileConfigor{
		rctx:    ctx,
		rcancel: cancel,
	}

	if _, err := f.loadConfigLocked(u.Path); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *fileConfigor) loadConfigLocked(fpath string) (payload []byte, err error) {
	payload, err = os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	if fpath == "" {
		return nil, errors.New("invalid file path")
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.fpath, f.payload = fpath, payload
	return
}

func (f *fileConfigor) Load() ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.payload, nil
}

// we should only care about write event
func (f *fileConfigor) Monitor(cb nacs.OnChange) (context.CancelFunc, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create watcher")
	}

	if err := watcher.Add(f.fpath); err != nil {
		return nil, errors.Wrap(err, "failed to watch file")
	}

	if cb == nil {
		cb = func(err error, a ...any) {}
	}

	ctx, cancel := context.WithCancel(f.rctx)
	go func() {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					payload, err := f.loadConfigLocked(event.Name)
					cb(err, payload, event.Name)
				}

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return cancel, nil
}

func (f *fileConfigor) Close() error {
	f.rcancel()
	return nil
}
