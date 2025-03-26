package gogen

import (
	_ "embed"
	"fmt"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
)

//go:embed service.tpl
var serviceTemplate string

func genService(dir, rootPkg string, api *spec.ApiSpec) error {
	middlewares := getMiddleware(api)

	middlewareVal := "[]gin.HandlerFunc{"
	for _, item := range middlewares {
		middlewareVal += fmt.Sprintf("\nmiddleware.New%s(),", item)
	}
	middlewareVal += "\n\t}"

	imports := ""
	if len(middlewares) > 0 {
		imports += "\n\t\"" + pathx.JoinPackages(rootPkg, middlewareDir) + "\""
		imports += "\n\t\"" + pathx.JoinPackages(rootPkg, handlerDir) + "\""
	}

	return genFile(fileGenConfig{
		rootpath:         dir,
		relativepath:     serviceDir,
		filename:         "service.go",
		templateName:     "serviceTemplate",
		category:         category,
		templateFileName: serviceTemplateFile,
		builtinTemplate:  serviceTemplate,
		data: map[string]string{
			"imports":     imports,
			"middlewares": middlewareVal,
		},
	})
}
