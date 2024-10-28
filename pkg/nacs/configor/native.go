package configor

import (
	"context"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FileConfigorBuilder struct{}

func (FileConfigorBuilder) Build() Configor { return NewFileConfigor() }

func init() {
	Register("native", FileConfigorBuilder{})
}

type FileConfigor struct{}

func NewFileConfigor() Configor {
	return &FileConfigor{}
}

func (fc *FileConfigor) Watch(ctx context.Context, handler Handler, paths ...string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for _, path := range paths {
		if err := watcher.Add(path); err != nil {
			watcher.Close()
			return err
		}
	}

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
				fc.handleEvent(event, handler)
			case err, ok := <-watcher.Errors:
				if ok {
					handler.OnChange(SYS, Config{}, err)
				}
			}
		}
	}()
	return nil
}

func (fc *FileConfigor) handleEvent(event fsnotify.Event, handler Handler) {
	var evt Event
	var content []byte
	var err error
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		evt = ADD
		content, err = os.ReadFile(event.Name)
	case event.Op&fsnotify.Write == fsnotify.Write:
		evt = CHG
		content, err = os.ReadFile(event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		evt = DEL
	default:
		return
	}
	handler.OnChange(evt, Config{event.Name, content}, err)
}

func (fc *FileConfigor) Load(ctx context.Context, paths ...string) (map[string][]byte, error) {
	m := make(map[string][]byte, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		m[path] = content
	}

	return m, nil
}
