package gogen

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
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
		filename:         stringx.ToSnake(handler) + "_test.go",
		templateName:     "handlerTestTemplate",
		category:         category,
		templateFileName: handlerTestTemplateFile,
		builtinTemplate:  handlerTestTemplate,
		data: map[string]any{
			"PkgName":      pkgName,
			"imports":      genHandlerTestImports(route, rootPkg),
			"HandlerName":  handler,
			"RequestType":  util.Title(route.RequestTypeName()),
			"ResponseType": respType,
			"HasResponse":  len(route.ResponseTypeName()) > 0,
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

func genHandlerTestImports(route spec.Route, parentPkg string) string {
	imports := []string{}
	if len(route.RequestTypeName()) > 0 || len(route.ResponseTypeName()) > 0 {
		imports = append(imports, fmt.Sprintf("\"%s\"", pathx.JoinPackages(parentPkg, typesDir)))
	}

	return strings.Join(imports, "\n\t")
}
