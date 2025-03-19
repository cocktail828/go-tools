package gogen

import (
	_ "embed"
	"fmt"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
)

func init() { Register(TypeMain, &mainGenerater{}) }

type mainGenerater struct {
	*spec.ApiSpec
}

func (g *mainGenerater) PkgName() string              { return "main" }
func (g *mainGenerater) RelativePath() string         { return "." }
func (g *mainGenerater) TemplateFile() string         { return "main.tpl" }
func (g *mainGenerater) Init(api *spec.ApiSpec) error { g.ApiSpec = api; return nil }
func (g *mainGenerater) Export() Export               { return Export{} }

func (g *mainGenerater) Gen(fm FileMeta) Render {
	// filename := strings.ToLower(g.Service.Name)

	// if strings.HasSuffix(filename, "-api") {
	// 	filename = strings.ReplaceAll(filename, "-api", "")
	// }

	return FileRender{
		rootpath:         fm.RootPath,
		relativepath:     g.RelativePath(),
		filename:         "main.go",
		templateFileName: g.TemplateFile(),
		data: map[string]string{
			"imports":     fmt.Sprintf("%q\n", pathx.JoinPackages(fm.Mod, fm.Lookup(TypeService).PkgName())),
		},
	}
}
