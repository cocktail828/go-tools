package gogen

import (
	"fmt"

	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
)

const (
	category                = "api"
	configTemplateFile      = "config.tpl"
	contextTemplateFile     = "context.tpl"
	etcTemplateFile         = "etc.tpl"
	handlerTemplateFile     = "handler.tpl"
	handlerTestTemplateFile = "handler_test.tpl"
	logicTemplateFile       = "logic.tpl"
	logicTestTemplateFile   = "logic_test.tpl"
	mainTemplateFile        = "main.tpl"
	middlewareImplementFile = "middleware.tpl"
	routesTemplateFile      = "routes.tpl"
	typesTemplateFile       = "types.tpl"
)

var templates = map[string]string{
	contextTemplateFile:     contextTemplate,
	handlerTemplateFile:     handlerTemplate,
	handlerTestTemplateFile: handlerTestTemplate,
	logicTemplateFile:       logicTemplate,
	logicTestTemplateFile:   logicTestTemplate,
	mainTemplateFile:        mainTemplate,
	middlewareImplementFile: middlewareImplement,
	routesTemplateFile:      routesTemplate,
	typesTemplateFile:       typesTemplate,
}

// Category returns the category of the api files.
func Category() string {
	return category
}

// Clean cleans the generated deployment files.
func Clean() error {
	return pathx.Clean(category)
}

// GenTemplates generates api template files.
func GenTemplates() error {
	return pathx.InitTemplates(category, templates)
}

// RevertTemplate reverts the given template file to the default value.
func RevertTemplate(name string) error {
	content, ok := templates[name]
	if !ok {
		return fmt.Errorf("%s: no such file name", name)
	}
	return pathx.CreateTemplate(category, name, content)
}

// Update updates the template files to the templates built in current goctl.
func Update() error {
	err := Clean()
	if err != nil {
		return err
	}

	return pathx.InitTemplates(category, templates)
}
