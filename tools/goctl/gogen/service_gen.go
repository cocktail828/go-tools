package gogen

import (
	_ "embed"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/z/stringx"
)

func init() { Register(TypeService, &serviceHandler{}) }

type serviceHandler struct {
	*spec.ApiSpec
}

func (g *serviceHandler) PkgName() string              { return "service" }
func (g *serviceHandler) RelativePath() string         { return "service" }
func (g *serviceHandler) TemplateFile() string         { return "service.tpl" }
func (g *serviceHandler) Init(api *spec.ApiSpec) error { g.ApiSpec = api; return nil }
func (g *serviceHandler) Export() Export               { return Export{} }
func (g *serviceHandler) Gen(fm FileMeta) Render {
	middlewares := fm.Lookup(TypeMiddleware).Export().Funcs

	mws := stringx.Array{}
	mws.WriteString("[]gin.HandlerFunc{")
	for _, m := range middlewares {
		mws.WriteStringf("middleware.New%s(),", m)
	}
	mws.WriteString("}")

	pkgs := stringx.Array{}
	if len(middlewares) > 0 {
		pkgs.WriteStringf("%q", pathx.JoinPackages(fm.Mod, fm.Lookup(TypeMiddleware).PkgName()))
		pkgs.WriteStringf("%q", pathx.JoinPackages(fm.Mod, fm.Lookup(TypeRoute).PkgName()))
	}

	return FileRender{
		rootpath:         fm.RootPath,
		relativepath:     g.RelativePath(),
		filename:         "service.go",
		templateFileName: g.TemplateFile(),
		data: map[string]string{
			"imports":     pkgs.Uniq().Join("\n"),
			"middlewares": mws.Join("\n"),
			"route":       fm.Lookup(TypeRoute).PkgName(),
		},
	}
}
