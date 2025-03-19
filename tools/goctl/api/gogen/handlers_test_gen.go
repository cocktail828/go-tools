package gogen

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/api/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/stringx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/util"
)

//go:embed handler_test.tpl
var handlerTestTemplate string

func genHandlerTest(dir, rootPkg string, group spec.Group, route spec.Route) error {
	handler := getHandlerName(route)
	handlerPath := getHandlerFolderPath(group, route)
	pkgName := handlerPath[strings.LastIndex(handlerPath, "/")+1:]
	logicName := defaultLogicPackage
	if handlerPath != handlerDir {
		handler = stringx.Title(handler)
		logicName = pkgName
	}

	filename := stringx.ToSnake(handler)
	return genFile(fileGenConfig{
		rootpath:        dir,
		relativepath:    getHandlerFolderPath(group, route),
		filename:        filename + "_test.go",
		templateName:    "handlerTestTemplate",
		category:        category,
		templateFile:    handlerTestTemplateFile,
		builtinTemplate: handlerTestTemplate,
		data: map[string]any{
			"PkgName":      pkgName,
			"imports":      genHandlerTestImports(group, route, rootPkg),
			"HandlerName":  handler + "Handler",
			"RequestType":  util.Title(route.RequestTypeName()),
			"ResponseType": util.Title(route.ResponseTypeName()),
			"LogicName":    logicName,
			"LogicType":    stringx.Title(getLogicName(route)),
			"Call":         stringx.Title(strings.TrimSuffix(handler, "Handler")),
			"HasResp":      len(route.ResponseTypeName()) > 0,
			"HasRequest":   len(route.RequestTypeName()) > 0,
			"HasDoc":       len(route.JoinedDoc()) > 0,
			"Doc":          getDoc(route.JoinedDoc()),
		},
	})
}

func genHandlersTest(dir, rootPkg string, api *spec.ApiSpec) error {
	for _, group := range api.Service.Groups {
		for _, route := range group.Routes {
			if err := genHandlerTest(dir, rootPkg, group, route); err != nil {
				return err
			}
		}
	}

	return nil
}

func genHandlerTestImports(group spec.Group, route spec.Route, parentPkg string) string {
	imports := []string{}
	if len(route.RequestTypeName()) > 0 {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", pathx.JoinPackages(parentPkg, typesDir)))
	}

	return strings.Join(imports, "\n\t")
}
