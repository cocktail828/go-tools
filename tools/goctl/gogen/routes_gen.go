package gogen

import (
	"path"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/version"
	"github.com/cocktail828/go-tools/z/stringx"
)

const (
	routesFilename = "routes"
)

type group struct {
	name   string // group name
	routes []route
	prefix string
}

type route struct {
	spec.Route
	method  string
	path    string
	handler string
	doc     string
}

func (grp group) ToRouteStr() string {
	sa := stringx.Array{}
	if grp.prefix != "" {
		gname := "group"
		if grp.name != "" {
			gname = stringx.Untitle(stringx.ToCamel(grp.name) + "Group")
		}

		sa.WriteStringf("%s := g.Group(%q)", gname, grp.prefix)
		sa.WriteString("{")
		for _, r := range grp.routes {
			sa.WriteStringf("%s.Handle(%s, %q, %s.%sHandler(m.Timeout, m.Logger, m.Meta))", gname, r.method, r.path, grp.name, stringx.Title(r.handler))
		}
		sa.WriteString("}")
	} else {
		for _, r := range grp.routes {
			sa.WriteStringf("g.Handle(%s, %q, %sHandler(m.Timeout, m.Logger, m.Meta))", r.method, r.path, stringx.Untitle(r.handler))
		}
	}

	return sa.Join("\n")
}

func toGroups(api *spec.ApiSpec) ([]group, error) {
	var groups []group

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

	for _, g := range api.Service.Groups {
		var grp group
		for _, rt := range g.Routes {
			handler := rt.Handler
			if folder := rt.GetAnnotation(groupProperty); len(folder) > 0 {
				grp.name = stringx.ToSnake(folder)
				handler = rt.Handler
			} else if folder = g.GetAnnotation(groupProperty); len(folder) > 0 {
				grp.name = stringx.ToSnake(folder)
				handler = rt.Handler
			}

			grp.routes = append(grp.routes, route{
				Route:   rt,
				method:  mapping[rt.Method],
				path:    rt.Path,
				handler: handler,
				doc:     rt.JoinedDoc(),
			})
		}

		prefix := g.GetAnnotation(spec.RoutePrefixKey)
		prefix = strings.ReplaceAll(prefix, `"`, "")
		prefix = strings.TrimSpace(prefix)
		if len(prefix) > 0 {
			prefix = path.Join("/", prefix)
			grp.prefix = prefix
		}
		groups = append(groups, grp)
	}

	return groups, nil
}

func init() { Register(TypeRoute, &routeHandler{}) }

type routeHandler struct {
	*spec.ApiSpec
}

func (g *routeHandler) PkgName() string              { return "route" }
func (g *routeHandler) RelativePath() string         { return "route" }
func (g *routeHandler) TemplateFile() string         { return "route.tpl" }
func (g *routeHandler) Init(api *spec.ApiSpec) error { g.ApiSpec = api; return nil }
func (g *routeHandler) Export() Export               { return Export{} }
func (g *routeHandler) Gen(fm FileMeta) Render {
	groups, err := toGroups(g.ApiSpec)
	if err != nil {
		return ErrRender{err}
	}

	mr := MultiRender{}
	routeStr := stringx.Array{}
	imports := stringx.Array{}
	for _, grp := range groups {
		routeStr.WriteString(grp.ToRouteStr())
		if grp.name != "" {
			imports.WriteStringf("%q", pathx.JoinPackages(fm.Mod, g.PkgName(), stringx.ToSnake(grp.name)))
		}

		for _, rt := range grp.routes {
			h := handlerGenerater{
				route:         rt,
				pkgName:       grp.name,
				releativePath: path.Join(g.RelativePath(), grp.name),
			}
			mr = append(mr, h.gen(fm, g.PkgName()))
			mr = append(mr, h.genTest(fm, g.PkgName()))
		}
	}

	middlewares := fm.Lookup(TypeMiddleware).Export().Funcs
	mws := stringx.Array{}
	mws.WriteStringf("")
	for _, m := range middlewares {
		mws.WriteStringf("middleware.New%s(m.Meta),", m)
	}
	mws.WriteStringf("")

	if len(middlewares) > 0 {
		imports.WriteStringf("%q", pathx.JoinPackages(fm.Mod, fm.Lookup(TypeMiddleware).PkgName()))
	}

	mr = append(mr, FileRender{
		rootpath:         fm.RootPath,
		relativepath:     g.RelativePath(),
		filename:         stringx.ToSnake(routesFilename) + ".go",
		templateFileName: g.TemplateFile(),
		data: map[string]any{
			"pkgName":    g.PkgName(),
			"imports":    imports.Uniq().Join("\n"),
			"middleware": mws.Join("\n"),
			"routes":     strings.TrimSpace(routeStr.Join("\n")),
			"version":    version.BuildVersion,
		},
	})
	return mr
}

type handlerGenerater struct {
	route
	pkgName       string
	releativePath string
}

func (g *handlerGenerater) gen(fm FileMeta, dftPkgName string) Render {
	respType := ""
	if resp := g.ResponseType; resp != nil {
		respType = "*" + fm.Lookup(TypeModel).PkgName() + "." + stringx.Title(g.ResponseTypeName())
		if tp, ok := resp.(spec.ArrayType); ok {
			respType = "[]" + fm.Lookup(TypeModel).PkgName() + "." + tp.Value.Name()
		}
	}

	imports := stringx.Array{}
	if len(g.RequestTypeName()) > 0 || len(g.ResponseTypeName()) > 0 {
		imports.WriteStringf("%q", pathx.JoinPackages(fm.Mod, fm.Lookup(TypeModel).PkgName()))
	}

	pkgName := g.pkgName
	if pkgName == "" {
		pkgName = dftPkgName
		g.handler = stringx.Untitle(g.handler)
	}else{
		g.handler = stringx.Title(g.handler)
	}

	return FileRender{
		rootpath:         fm.RootPath,
		relativepath:     g.releativePath,
		filename:         stringx.ToSnake(g.handler) + ".go",
		templateFileName: "handler.tpl",
		data: map[string]any{
			"pkgName":      pkgName,
			"imports":      imports.Join("\n"),
			"handler":      g.handler,
			"requestType":  stringx.Title(g.RequestTypeName()),
			"responseType": respType,
		},
	}
}

func (g *handlerGenerater) genTest(fm FileMeta, dftPkgName string) Render {
	respType := ""
	if resp := g.ResponseType; resp != nil {
		respType = "*" + fm.Lookup(TypeModel).PkgName() + "." + stringx.Title(g.ResponseTypeName())
		if tp, ok := resp.(spec.ArrayType); ok {
			respType = "[]" + fm.Lookup(TypeModel).PkgName() + "." + tp.Value.Name()
		}
	}

	imports := stringx.Array{}
	if len(g.RequestTypeName()) > 0 || len(g.ResponseTypeName()) > 0 {
		imports.WriteStringf("%q", pathx.JoinPackages(fm.Mod, fm.Lookup(TypeModel).PkgName()))
	}

	pkgName := g.pkgName
	if pkgName == "" {
		pkgName = dftPkgName
		g.handler = stringx.Untitle(g.handler)
	}

	return FileRender{
		rootpath:         fm.RootPath,
		relativepath:     g.releativePath,
		filename:         stringx.ToSnake(g.handler) + "_test.go",
		templateFileName: "handler_test.tpl",
		data: map[string]any{
			"pkgName":      pkgName,
			"imports":      imports.Join("\n"),
			"handler":      g.handler,
			"requestType":  stringx.Title(g.RequestTypeName()),
			"responseType": respType,
			// "doc":          route.JoinedDoc(),
		},
	}
}
