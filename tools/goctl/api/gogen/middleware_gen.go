package gogen

import (
	_ "embed"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/stringx"
)

//go:embed middleware.tpl
var middlewareImplement string

func genMiddleware(dir string, api *spec.ApiSpec) error {
	middlewares := getMiddleware(api)
	for _, item := range middlewares {
		err := genFile(fileGenConfig{
			rootpath:         dir,
			relativepath:     middlewareDir,
			filename:         stringx.ToSnake(item) + ".go",
			templateName:     "contextTemplate",
			category:         category,
			templateFileName: middlewareImplementFile,
			builtinTemplate:  middlewareImplement,
			data: map[string]string{
				"name": stringx.Title(item),
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
