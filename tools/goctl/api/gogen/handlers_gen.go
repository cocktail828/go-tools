package gogen

import (
	_ "embed"
	"fmt"
	"path"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/api/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/stringx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/util"
	"github.com/cocktail828/go-tools/z"
)

const defaultLogicPackage = "logic"

//go:embed handler.tpl
var handlerTemplate string

func genHandler(dir, rootPkg string, group spec.Group, route spec.Route) error {
	handler := getHandlerName(route)
	handlerPath := getHandlerFolderPath(group, route)
	pkgName := handlerPath[strings.LastIndex(handlerPath, "/")+1:]
	logicName := defaultLogicPackage
	if handlerPath != handlerDir {
		handler = stringx.Title(handler)
		logicName = pkgName
	}

	return genFile(fileGenConfig{
		rootpath:        dir,
		relativepath:    getHandlerFolderPath(group, route),
		filename:        stringx.ToSnake(handler) + ".go",
		templateName:    "handlerTemplate",
		category:        category,
		templateFile:    handlerTemplateFile,
		builtinTemplate: handlerTemplate,
		data: map[string]any{
			"PkgName":        pkgName,
			"ImportPackages": genHandlerImports(group, route, rootPkg),
			"HandlerName":    handler + "Handler",
			"RequestType":    util.Title(route.RequestTypeName()),
			"LogicName":      logicName,
			"LogicType":      stringx.Title(getLogicName(route)) + "Logic",
			"Call":           stringx.Title(strings.TrimSuffix(handler, "Handler")),
			"HasResp":        len(route.ResponseTypeName()) > 0,
			"HasRequest":     len(route.RequestTypeName()) > 0,
			"HasDoc":         len(route.JoinedDoc()) > 0,
			"Doc":            getDoc(route.JoinedDoc()),
		},
	})
}

func genHandlers(dir, rootPkg string, api *spec.ApiSpec) error {
	for _, group := range api.Service.Groups {
		for _, route := range group.Routes {
			if err := genHandler(dir, rootPkg, group, route); err != nil {
				return err
			}
		}
	}
	return nil
}

func genHandlerImports(group spec.Group, route spec.Route, parentPkg string) string {
	imports := []string{
		fmt.Sprintf("%q", pathx.JoinPackages(parentPkg, getLogicFolderPath(group, route))),
	}
	if len(route.RequestTypeName()) > 0 {
		imports = append(imports, fmt.Sprintf("%q\n", pathx.JoinPackages(parentPkg, typesDir)))
	}

	return strings.Join(imports, "\n\t")
}

func getHandlerBaseName(route spec.Route) (string, error) {
	handler := route.Handler
	handler = strings.TrimSpace(handler)
	handler = strings.TrimSuffix(handler, "handler")
	handler = strings.TrimSuffix(handler, "Handler")

	return handler, nil
}

func getHandlerFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(groupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(groupProperty)
		if len(folder) == 0 {
			return handlerDir
		}
	}

	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")

	return path.Join(handlerDir, folder)
}

func getHandlerName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	z.Must(err)
	return handler
}

func getLogicName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	z.Must(err)
	return handler
}
