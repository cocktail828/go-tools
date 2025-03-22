package gogen

import (
	_ "embed"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/api/parser/g4/api"
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
	if handlerPath != handlerDir {
		handler = stringx.Title(handler)
	}

	var responseString string
	var returnString string
	var requestString string
	if len(route.ResponseTypeName()) > 0 {
		resp := responseGoTypeName(route, typesPacket)
		responseString = "(resp " + resp + ", err error)"
		returnString = "return"
	} else {
		responseString = "error"
		returnString = "return nil"
	}
	if len(route.RequestTypeName()) > 0 {
		requestString = "req *" + requestGoTypeName(route, typesPacket)
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
			"PkgName":      pkgName,
			"imports":      genHandlerImports(group, route, rootPkg),
			"HandlerName":  handler + "Handler",
			"RequestType":  util.Title(route.RequestTypeName()),
			"Call":         stringx.Title(strings.TrimSuffix(handler, "Handler")),
			"HasResp":      len(route.ResponseTypeName()) > 0,
			"HasRequest":   len(route.RequestTypeName()) > 0,
			"HasDoc":       len(route.JoinedDoc()) > 0,
			"Doc":          getDoc(route.JoinedDoc()),
			"ResponseType": responseString,
			"ReturnString": returnString,
			"Request":      requestString,
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

func getLogicName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	z.Must(err)
	return handler
}

func onlyPrimitiveTypes(val string) bool {
	fields := strings.FieldsFunc(val, func(r rune) bool {
		return r == '[' || r == ']' || r == ' '
	})

	for _, field := range fields {
		if field == "map" {
			continue
		}
		// ignore array dimension number, like [5]int
		if _, err := strconv.Atoi(field); err == nil {
			continue
		}
		if !api.IsBasicType(field) {
			return false
		}
	}

	return true
}

func shallImportTypesPackage(route spec.Route) bool {
	if len(route.RequestTypeName()) > 0 {
		return true
	}

	respTypeName := route.ResponseTypeName()
	if len(respTypeName) == 0 {
		return false
	}

	if onlyPrimitiveTypes(respTypeName) {
		return false
	}

	return true
}
