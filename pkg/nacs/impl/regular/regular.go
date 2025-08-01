package regular

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/cocktail828/go-tools/pkg/nacs/configuration"
	"github.com/cocktail828/go-tools/z"
	"github.com/pkg/errors"
	"gopkg.in/fsnotify.v1"
)

type fileConfigor struct {
	runningCtx context.Context
	cancel     context.CancelFunc
	filters    []Filter
	root       string
	mu         sync.RWMutex
	configs    map[string][]byte // name -> payload
}

func NewFileConfigor(root string, filters ...Filter) (configuration.Configor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	fc := fileConfigor{
		runningCtx: ctx,
		cancel:     cancel,
		filters:    filters,
		root:       root,
		configs:    map[string][]byte{},
	}

	os.MkdirAll(root, os.ModePerm)
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		if d.IsDir() {
			return nil
		}

		return fc.loadConfigLocked(path, dirEntryImpl{d.Name()})
	}); err != nil {
		return nil, err
	}

	return &fc, nil
}

func (f *fileConfigor) loadConfigLocked(path string, d DirEntry) (err error) {
	for _, f := range f.filters {
		if f(d) {
			return
		}
	}

	var payload []byte
	payload, err = os.ReadFile(path)
	if err != nil {
		return
	}

	z.WithLock(&f.mu, func() {
		f.configs[d.Name()] = payload
	})

	return
}

func filename(cfg configuration.Config) string {
	return cfg.ID
}

func (f *fileConfigor) Get(cfg configuration.Config) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	value, ok := f.configs[filename(cfg)]
	if !ok {
		return nil, errors.Errorf("config fname %s not found", filename(cfg))
	}
	return value, nil
}

func (f *fileConfigor) Set(cfg configuration.Config, payload []byte) error {
	z.WithLock(&f.mu, func() {
		f.configs[filename(cfg)] = payload
	})
	os.WriteFile(path.Join(f.root, filename(cfg)), payload, os.ModePerm)
	return nil
}

func (f *fileConfigor) Delete(cfg configuration.Config) (err error) {
	z.WithLock(&f.mu, func() {
		delete(f.configs, filename(cfg))
	})
	os.Remove(path.Join(f.root, filename(cfg)))
	return
}

// we should only care about write event
func (f *fileConfigor) Monitor(cfg configuration.Config, listener configuration.Listener) (context.CancelFunc, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Errorf("failed to create watcher: %v", err)
	}

	if err := watcher.Add(f.root); err != nil {
		return nil, errors.Errorf("failed to watch file: %v", err)
	}

	re, err := regexp.Compile(filename(cfg))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(f.runningCtx)
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

				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					fname := filepath.Base(event.Name)
					if !re.MatchString(fname) {
						continue
					}

					cfg := configuration.Config{ID: fname}
					oldval, _ := f.Get(cfg)
					if err := f.loadConfigLocked(event.Name, dirEntryImpl{fname}); err != nil {
						listener(cfg, nil, err)
						continue
					}

					current, _ := f.Get(cfg)
					if !bytes.Equal(oldval, current) {
						listener(cfg, current, nil)
					}
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
	f.cancel()
	return nil
}
