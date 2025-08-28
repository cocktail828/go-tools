package gen

import (
	"os"
	"path/filepath"
)

type Writer interface {
	Write(root string) error
}

type File struct {
	Path    string
	Name    string
	Payload string
}

func (f File) Write(root string) error {
	if f.Path != "" {
		os.MkdirAll(filepath.Join(root, f.Path), 0755)
	}

	return os.WriteFile(filepath.Join(root, f.Path, f.Name), []byte(f.Payload), 0644)
}

type MultiFile []File

func (mf MultiFile) Write(root string) error {
	for _, f := range mf {
		if err := f.Write(root); err != nil {
			return err
		}
	}

	return nil
}
