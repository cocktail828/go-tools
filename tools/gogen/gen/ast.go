package gen

import (
	"strings"

	"github.com/cocktail828/go-tools/tools/gogen/ast"
)

type DSLMeta struct {
	Syntax   string
	Project  string
	Services []serviceMeta
	Structs  []ast.StructDef
}

type serviceMeta struct {
	HasInterceptor bool
	Interceptors   []string
	Groups         []groupMeta
}

type groupMeta struct {
	Name         string // group name
	Prefix       string
	Interceptors []string
	Routes       []routeMeta
}

type routeMeta struct {
	HandlerName string
	Method      string
	Path        string
	Request     string
	Response    string
}

func serviceAst2Meta(svc ast.ServiceAST) serviceMeta {
	m := serviceMeta{
		Interceptors: titleSlice(strings.Title, svc.Interceptors...),
	}

	grpHasIncp := false
	for _, rg := range svc.Groups {
		grp := groupMeta{
			Name:         strings.ReplaceAll(rg.Prefix, "/", "_"),
			Prefix:       rg.Prefix,
			Interceptors: titleSlice(strings.Title, rg.Interceptors...),
		}

		if len(rg.Interceptors) > 0 {
			grpHasIncp = true
		}

		for _, rrt := range rg.Routes {
			if rrt.Path[0] != '/' {
				panic("invalid group prefix, should start with '/'")
			}

			grp.Routes = append(grp.Routes, routeMeta{
				HandlerName: strings.Title(rrt.HandlerName),
				Method:      strings.Title(rrt.Method),
				Path:        rrt.Path,
				Request:     rrt.Request,
				Response:    rrt.Response,
			})
		}
		m.Groups = append(m.Groups, grp)
	}
	m.HasInterceptor = len(svc.Interceptors) > 0 || grpHasIncp

	return m
}
