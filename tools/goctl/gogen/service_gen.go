package gogen

import (
	_ "embed"
	"fmt"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
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
	return FileRender{
		rootpath:         fm.RootPath,
		relativepath:     g.RelativePath(),
		filename:         "service.go",
		templateFileName: g.TemplateFile(),
		data: map[string]string{
			"imports":     fmt.Sprintf("%q", pathx.JoinPackages(fm.Mod, fm.Lookup(TypeRoute).PkgName())),
			"route":       fm.Lookup(TypeRoute).PkgName(),
		},
	}
}
