package gogen

import (
	_ "embed"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
)

func init() { Register(TypeMakefile, &makefileGenerater{}) }

type makefileGenerater struct {
	*spec.ApiSpec
}

func (g *makefileGenerater) PkgName() string              { return "" }
func (g *makefileGenerater) RelativePath() string         { return "." }
func (g *makefileGenerater) TemplateFile() string         { return "makefile.tpl" }
func (g *makefileGenerater) Init(api *spec.ApiSpec) error { g.ApiSpec = api; return nil }
func (g *makefileGenerater) Export() Export               { return Export{} }

func (g *makefileGenerater) Gen(fm FileMeta) Render {
	return FileRender{
		rootpath:         fm.RootPath,
		relativepath:     g.RelativePath(),
		filename:         "Makefile",
		templateFileName: g.TemplateFile(),
		data: map[string]string{
			"mod": fm.Mod,
		},
	}
}
