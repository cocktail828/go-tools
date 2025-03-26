package gogen

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/collection"
	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/stringx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/version"
)

//go:embed route.tpl
var routesTemplate string

const (
	routesFilename = "routes"
)

var mapping = map[string]string{
	"delete":  "http.MethodDelete",
	"get":     "http.MethodGet",
	"head":    "http.MethodHead",
	"post":    "http.MethodPost",
	"put":     "http.MethodPut",
	"patch":   "http.MethodPatch",
	"connect": "http.MethodConnect",
	"options": "http.MethodOptions",
	"trace":   "http.MethodTrace",
}

type (
	group struct {
		name   string
		routes []route
		prefix string
	}
	route struct {
		method  string
		path    string
		handler string
		doc     string
	}
)

func genRoutes(dir, rootPkg string, api *spec.ApiSpec) error {
	groups, err := getRoutes(api)
	if err != nil {
		return err
	}

	var builder strings.Builder
	for _, g := range groups {
		var gbuilder strings.Builder
		if g.prefix != "" {
			gname := "group"
			if g.name != "" {
				gname = stringx.Untitle(stringx.ToCamel(g.name) + "Group")
			}

			fmt.Fprintf(&gbuilder, "\n\t%s := g.Group(%q)\n\t{", gname, g.prefix)
			for _, r := range g.routes {
				fmt.Fprintf(&gbuilder, "\n\t\t%s.Handle(%s, %q, %s)", gname, r.method, r.path, r.handler)
			}
			fmt.Fprintf(&gbuilder, "\n\t}")
		} else {
			for _, r := range g.routes {
				fmt.Fprintf(&gbuilder, "\n\tg.Handle(%s, %q, %s)", r.method, r.path, r.handler)
			}
		}

		fmt.Fprintf(&builder, "\n\t%s\n", strings.TrimSpace(gbuilder.String()))
	}

	routeFilename := stringx.ToSnake(routesFilename) + ".go"
	filename := path.Join(dir, handlerDir, routeFilename)
	os.Remove(filename)

	return genFile(fileGenConfig{
		rootpath:         dir,
		relativepath:     handlerDir,
		filename:         routeFilename,
		templateName:     "routesTemplate",
		category:         category,
		templateFileName: routesTemplateFile,
		builtinTemplate:  routesTemplate,
		data: map[string]any{
			"imports": genRouteImports(rootPkg, api),
			"routes":  strings.TrimSpace(builder.String()),
			"version": version.BuildVersion,
		},
	})
}

func genRouteImports(parentPkg string, api *spec.ApiSpec) string {
	importSet := collection.NewSet()
	for _, group := range api.Service.Groups {
		for _, route := range group.Routes {
			folder := route.GetAnnotation(groupProperty)
			if len(folder) == 0 {
				folder = group.GetAnnotation(groupProperty)
				if len(folder) == 0 {
					continue
				}
			}
			importSet.AddStr(fmt.Sprintf("%q", pathx.JoinPackages(parentPkg, handlerDir, stringx.ToSnake(folder))))
		}
	}
	imports := importSet.KeysStr()
	sort.Strings(imports)
	return strings.Join(imports, "\n\t")
}

func getRoutes(api *spec.ApiSpec) ([]group, error) {
	var routes []group

	for _, g := range api.Service.Groups {
		var groupedRoutes group
		for _, r := range g.Routes {
			handler := getHandlerName(r) + "Handler(meta.Timeout, meta.Logger)"
			if folder := r.GetAnnotation(groupProperty); len(folder) > 0 {
				folder = stringx.ToSnake(folder)
				handler = toPrefix(folder) + "." + strings.ToUpper(handler[:1]) + handler[1:]
			} else if folder = g.GetAnnotation(groupProperty); len(folder) > 0 {
				folder = stringx.ToSnake(folder)
				groupedRoutes.name = folder
				handler = toPrefix(folder) + "." + strings.ToUpper(handler[:1]) + handler[1:]
			}
			groupedRoutes.routes = append(groupedRoutes.routes, route{
				method:  mapping[r.Method],
				path:    r.Path,
				handler: handler,
				doc:     r.JoinedDoc(),
			})
		}

		prefix := g.GetAnnotation(spec.RoutePrefixKey)
		prefix = strings.ReplaceAll(prefix, `"`, "")
		prefix = strings.TrimSpace(prefix)
		if len(prefix) > 0 {
			prefix = path.Join("/", prefix)
			groupedRoutes.prefix = prefix
		}
		routes = append(routes, groupedRoutes)
	}

	return routes, nil
}

func toPrefix(folder string) string {
	return strings.ReplaceAll(folder, "/", "")
}
