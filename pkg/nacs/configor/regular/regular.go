package regular

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/cocktail828/go-tools/z"
	"github.com/pkg/errors"
	"gopkg.in/fsnotify.v1"
)

var _ nacs.Configor = &FileConfigor{}

type FileConfigor struct {
	closed  atomic.Bool
	filters []Filter
	root    string
	watcher *fsnotify.Watcher
	mu      sync.RWMutex
	configs map[string][]byte
}

func NewFileConfigor(root string, filters ...Filter) (*FileConfigor, error) {
	fc := FileConfigor{
		filters: filters,
		root:    root,
		configs: map[string][]byte{},
	}

	os.MkdirAll(root, os.ModeDir)
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

func (f *FileConfigor) loadConfigLocked(path string, d DirEntry) (err error) {
	for _, f := range f.filters {
		if f(d) {
			return
		}
	}

	z.WithLock(&f.mu, func() {
		var payload []byte
		payload, err = os.ReadFile(path)
		if err != nil {
			return
		}
		f.configs[d.Name()] = payload
	})

	return
}

func (f *FileConfigor) GetConfig(fname string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	value, ok := f.configs[fname]
	if !ok {
		return nil, errors.Errorf("config fname %s not found", fname)
	}
	return value, nil
}

func (f *FileConfigor) SetConfig(fname string, payload []byte) error {
	z.WithLock(&f.mu, func() {
		f.configs[fname] = payload
	})
	os.WriteFile(path.Join(f.root, fname), payload, os.ModePerm)
	return nil
}

func (f *FileConfigor) DeleteConfig(fname string) (err error) {
	z.WithLock(&f.mu, func() {
		delete(f.configs, fname)
	})
	os.Remove(path.Join(f.root, fname))
	return
}

// we should only care about write event
func (f *FileConfigor) WatchConfig(listener nacs.ConfigListener) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Errorf("failed to create watcher: %v", err)
	}

	if err := watcher.Add(f.root); err != nil {
		return errors.Errorf("failed to watch file: %v", err)
	}
	f.watcher = watcher
	go func() {
		for {
			select {
			case event, ok := <-f.watcher.Events:
				if !ok {
					return
				}

				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					fname := filepath.Base(event.Name)
					oldval, _ := f.GetConfig(fname)
					if err := f.loadConfigLocked(event.Name, dirEntryImpl{fname}); err != nil {
						listener(event.Name, nil, err)
						continue
					}

					current, _ := f.GetConfig(fname)
					if !bytes.Equal(oldval, current) {
						listener(fname, current, nil)
					}
				}

			case _, ok := <-f.watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return nil
}

func (f *FileConfigor) Close() error {
	if f.closed.CompareAndSwap(false, true) {
		if f.watcher != nil {
			return f.watcher.Close()
		}
	}
	return nil
}
