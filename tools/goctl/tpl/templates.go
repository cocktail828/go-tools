package tpl

import (
	"path/filepath"

	"github.com/cocktail828/go-tools/tools/goctl/gogen"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	rpcgen "github.com/cocktail828/go-tools/tools/goctl/rpc/generator"
	"github.com/cocktail828/go-tools/xlog/colorful"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const templateParentPath = "/"

// genTemplates writes the latest template text into file which is not exists
func genTemplates(_ *cobra.Command, _ []string) error {
	path := varStringHome
	if len(path) != 0 {
		pathx.RegisterGoctlHome(path)
	}

	fns := []func() error{
		gogen.GenTemplates,
		rpcgen.GenTemplates,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	dir, err := pathx.GetTemplateDir(templateParentPath)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	colorful.Warnf("Templates are generated in %s, edit on your risk!\n", abs)
	return nil
}

// cleanTemplates deletes all templates
func cleanTemplates(_ *cobra.Command, _ []string) error {
	path := varStringHome
	if len(path) != 0 {
		pathx.RegisterGoctlHome(path)
	}

	fns := []func() error{
		gogen.Clean,
		rpcgen.Clean,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	colorful.Infoln("templates are cleaned!")
	return nil
}

// updateTemplates writes the latest template text into file,
// it will delete the older templates if there are exists
func updateTemplates(_ *cobra.Command, _ []string) (err error) {
	path := varStringHome
	category := varStringCategory
	if len(path) != 0 {
		pathx.RegisterGoctlHome(path)
	}

	defer func() {
		if err == nil {
			colorful.Errorf("%s template are update!", category)
		}
	}()
	switch category {
	case gogen.Category():
		return gogen.Update()
	case rpcgen.Category():
		return rpcgen.Update()
	default:
		err = errors.Errorf("unexpected category: %s", category)
		return
	}
}

// revertTemplates will overwrite the old template content with the new template
func revertTemplates(_ *cobra.Command, _ []string) (err error) {
	path := varStringHome
	category := varStringCategory
	filename := varStringName
	if len(path) != 0 {
		pathx.RegisterGoctlHome(path)
	}

	defer func() {
		if err == nil {
			colorful.Errorf("%s template are reverted!", filename)
		}
	}()
	switch category {
	case gogen.Category():
		return gogen.RevertTemplate(filename)
	case rpcgen.Category():
		return rpcgen.RevertTemplate(filename)
	default:
		err = errors.Errorf("unexpected category: %s", category)
		return
	}
}
