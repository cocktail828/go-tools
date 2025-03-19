package gogen

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var varStrHome, varStrMod, varStrAPI, varStrDir string
	cmd := &cobra.Command{
		Use:   "go",
		Short: "Generate go project",
		Run: func(_ *cobra.Command, _ []string) {
			genGoProject(varStrMod, varStrDir, varStrHome, varStrAPI)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&varStrHome, "home", "", "the default api file home")
	flags.StringVar(&varStrMod, "mod", "demo", "the default module name")
	flags.StringVar(&varStrDir, "dir", "", "the target source file directory")
	flags.StringVar(&varStrAPI, "api", "", "the api file")
	cmd.MarkFlagRequired("dir")
	cmd.MarkFlagRequired("api")

	return cmd
}

func Category() string { return "api" }

// GenTemplates generates api template files.
func GenTemplates() error {
	return nil
	// return pathx.InitTemplates(tpl.category, tpl.buildin)
}

// Clean cleans the generated deployment files.
func Clean() error {
	return nil
	// return pathx.Clean(tpl.category)
}

// RevertTemplate reverts the given template file to the default value.
func RevertTemplate(name string) error {
	return nil
	// content, ok := templates[name]
	// if !ok {
	// 	return errors.Errorf("%s: no such file name", name)
	// }
	// return pathx.CreateTemplate(tpl.category, name, content)
}

// Update updates the template files to the templates built in current goctl.
func Update() error {
	return nil
	// if err := tpl.Clean(); err != nil {
	// 	return err
	// }

	// return pathx.InitTemplates(tpl.category, templates)
}
