package native

import (
	"context"
	stderr "errors"
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
	configs sync.Map
}

// native://localhost?file=path1
func NewNativeConfigor(uri string) (nacs.Configor, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	fc := &fileConfigor{
		rctx:    ctx,
		rcancel: cancel,
	}

	if _, err := fc.loadConfigLocked(u.Query().Get("file")); err != nil {
		return nil, err
	}
	return fc, nil
}

func (f *fileConfigor) loadConfigLocked(fpath string) ([]byte, error) {
	payload, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	f.configs.Store(fpath, payload)
	return payload, nil
}

type nativeLoadOpt struct {
	fpath string
}

func FileName(v string) nacs.LoadOpt {
	return func(o any) {
		if f, ok := o.(*nativeLoadOpt); ok {
			f.fpath = v
		}
	}
}

func (f *fileConfigor) Load(opts ...nacs.LoadOpt) ([]byte, error) {
	var ro nativeLoadOpt
	for _, o := range opts {
		o(&ro)
	}

	if ro.fpath == "" {
		return nil, errors.New("invalid get opt")
	}

	value, ok := f.configs.Load(ro.fpath)
	if !ok {
		return nil, errors.Errorf("config %s not found", ro.fpath)
	}

	return value.([]byte), nil
}

// we should only care about write event
func (f *fileConfigor) Monitor(cb nacs.OnChange, opts ...nacs.MonitorOpt) (context.CancelFunc, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Errorf("failed to create watcher: %v", err)
	}

	errs := []error{}
	f.configs.Range(func(key, _ any) bool {
		if err := watcher.Add(key.(string)); err != nil {
			errs = append(errs, err)
		}
		return true
	})
	if err := stderr.Join(errs...); err != nil {
		return nil, errors.Errorf("failed to watch file: %v", err)
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
