package gogen

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
)

//go:embed main.tpl
var mainTemplate string

func genMain(dir, rootPkg string, api *spec.ApiSpec) error {
	filename := strings.ToLower(api.Service.Name)

	if strings.HasSuffix(filename, "-api") {
		filename = strings.ReplaceAll(filename, "-api", "")
	}

	return genFile(fileGenConfig{
		rootpath:         dir,
		relativepath:     "",
		filename:         filename + ".go",
		templateName:     "mainTemplate",
		category:         category,
		templateFileName: mainTemplateFile,
		builtinTemplate:  mainTemplate,
		data: map[string]string{
			"imports":     genMainImports(rootPkg),
			"serviceName": filename,
		},
	})
}

func genMainImports(parentPkg string) string {
	return fmt.Sprintf("%q\n", pathx.JoinPackages(parentPkg, serviceDir))
}
