package gogen

import (
	"embed"
	"os"
	"path"

	"github.com/pkg/errors"
)

var (
	//go:embed *.tpl
	buildinTmpl       embed.FS // buildin templates
	goctlTemplateRoot = os.Getenv("HOME") + "/.goctl/api"
)

type Template struct{}

func (tpl *Template) Load(relativepath, fname string) ([]byte, error) {
	body, err := os.ReadFile(path.Join(relativepath, fname))
	if err == nil {
		return body, nil
	}

	// fallback to cache dir
	body, err = os.ReadFile(path.Join(goctlTemplateRoot, fname))
	if err == nil {
		return body, nil
	}

	// fallback to buildin
	entries, err := buildinTmpl.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || entry.Name() != fname {
			continue
		}

		return buildinTmpl.ReadFile(entry.Name())
	}
	return nil, errors.Errorf("%q cannot find the specified template", fname)
}
