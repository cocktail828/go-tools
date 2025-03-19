package gogen

import (
	_ "embed"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/z/stringx"
	"github.com/samber/lo"
)

func init() { Register(TypeMiddleware, &middlewareGenerater{}) }

type middlewareGenerater struct {
	*spec.ApiSpec
	exp Export
}

func (g *middlewareGenerater) PkgName() string      { return "middleware" }
func (g *middlewareGenerater) RelativePath() string { return "middleware" }
func (g *middlewareGenerater) TemplateFile() string { return "middleware.tpl" }
func (g *middlewareGenerater) Init(api *spec.ApiSpec) error {
	g.ApiSpec = api
	for _, grp := range g.Service.Groups {
		middleware := grp.GetAnnotation("middleware")
		if len(middleware) > 0 {
			for _, item := range strings.Split(middleware, ",") {
				g.exp.Funcs = append(g.exp.Funcs, strings.TrimSpace(item))
			}
		}
	}

	g.exp.Funcs = lo.Uniq(g.exp.Funcs)
	return nil
}

func (g *middlewareGenerater) Export() Export { return g.exp }

func (g *middlewareGenerater) Gen(fm FileMeta) Render {
	mr := MultiRender{}
	for _, item := range g.Export().Funcs {
		mr = append(mr, FileRender{
			rootpath:         fm.RootPath,
			relativepath:     g.RelativePath(),
			filename:         stringx.ToSnake(item) + ".go",
			templateFileName: g.TemplateFile(),
			data: map[string]string{
				"name": stringx.Title(item),
			},
		})
	}

	return mr
}
