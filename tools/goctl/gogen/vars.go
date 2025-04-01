package gogen

import (
	_ "embed"
	"fmt"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/version"
)

func init() { Register(TypeVars, &varsGenerater{}) }

type varsGenerater struct {
	*spec.ApiSpec
}

func (g *varsGenerater) PkgName() string              { return "" }
func (g *varsGenerater) RelativePath() string         { return "vars" }
func (g *varsGenerater) TemplateFile() string         { return "" }
func (g *varsGenerater) Init(api *spec.ApiSpec) error { g.ApiSpec = api; return nil }
func (g *varsGenerater) Export() Export               { return Export{} }

func (g *varsGenerater) Gen(fm FileMeta) Render {
	return NopRender{
		rootpath:     fm.RootPath,
		relativepath: g.RelativePath(),
		filename:     "version.go",
		data: []byte(fmt.Sprintf(`
// Code generated by goctl. DO NOT EDIT.
// goctl %s

package vars

// The following variables will be injected during build time.
// Do not modify them manually as they will be overwritten by the build system.
var (
	GitTag     = "" // The current Git tag (version) of the codebase
	CommitLog  = "" // The latest Git commit hash and message
	BuildTime  = "" // Timestamp when the binary was built (format: YYYY-MM-DD HH:MM:SS)
	GoVersion  = "" // Go compiler version used for building
	AppVersion = "" // Version of the application (may combine GitTag and other metadata)
)
`, version.BuildVersion)),
	}
}
