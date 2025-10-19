package native

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
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

// file:///tmp/test_config.txt
// file://./test_config.txt
func NewFileConfigor(u *url.URL) (nacs.Configor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	f := &fileConfigor{
		rctx:    ctx,
		rcancel: cancel,
		fpath:   filepath.Join(u.Host, u.Path),
	}

	if _, err := f.loadConfigLocked(f.fpath); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *fileConfigor) loadConfigLocked(fpath string) (payload []byte, err error) {
	if fpath == "" {
		return nil, errors.New("empty file path")
	}

	payload, err = os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.payload = payload
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
		return nil, errors.Wrap(err, "failed to create fsnotify watcher")
	}

	if err := watcher.Add(f.fpath); err != nil {
		return nil, errors.Wrapf(err, "failed to watch file %s", f.fpath)
	}

	if cb == nil {
		cb = func(name string, payload []byte, err error) {}
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
					cb(event.Name, payload, err)
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
