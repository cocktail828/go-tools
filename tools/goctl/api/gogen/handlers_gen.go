package gogen

import (
	_ "embed"
	"fmt"
	"path"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/stringx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/util"
	"github.com/cocktail828/go-tools/z"
)

//go:embed handler.tpl
var handlerTemplate string

func genHandler(dir, rootPkg string, group spec.Group, route spec.Route) error {
	handler := getHandlerName(route)
	handlerPath := getHandlerFolderPath(group, route)
	if handlerPath != handlerDir {
		handler = stringx.Title(handler)
	}

	respType := "*" + typesPacket + "." + util.Title(route.ResponseTypeName())
	if resp := route.ResponseType; resp != nil {
		if tp, ok := resp.(spec.ArrayType); ok {
			respType = "[]" + typesPacket + "." + tp.Value.Name()
		}
	}

	return genFile(fileGenConfig{
		rootpath:         dir,
		relativepath:     getHandlerFolderPath(group, route),
		filename:         stringx.ToSnake(handler) + ".go",
		templateName:     "handlerTemplate",
		category:         category,
		templateFileName: handlerTemplateFile,
		builtinTemplate:  handlerTemplate,
		data: map[string]any{
			"PkgName":      handlerPath[strings.LastIndex(handlerPath, "/")+1:],
			"imports":      genHandlerImports(route, rootPkg),
			"HandlerName":  handler,
			"RequestType":  util.Title(route.RequestTypeName()),
			"ResponseType": respType,
			"HasRequest":   len(route.RequestTypeName()) > 0,
			"HasResponse":  len(route.ResponseTypeName()) > 0,
			"Doc":          getDoc(route.JoinedDoc()),
			"HasDoc":       len(route.JoinedDoc()) > 0,
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

func genHandlerImports(route spec.Route, parentPkg string) string {
	imports := []string{}
	if len(route.RequestTypeName()) > 0 || len(route.ResponseTypeName()) > 0 {
		imports = append(imports, fmt.Sprintf("%q", pathx.JoinPackages(parentPkg, typesDir)))
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
