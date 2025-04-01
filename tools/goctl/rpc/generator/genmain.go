package generator

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/util"
	"github.com/cocktail828/go-tools/tools/goctl/rpc/parser"
	"github.com/cocktail828/go-tools/z/stringx"
)

//go:embed main.tpl
var mainTemplate string

type MainServiceTemplateData struct {
	GRPCService string
	Service     string
	ServerPkg   string
	Pkg         string
}

// GenMain generates the main file of the rpc service, which is an rpc service program call entry
func (g *Generator) GenMain(ctx DirContext, proto parser.Proto, c *ZRpcContext) error {
	fileName := filepath.Join(ctx.GetMain().Filename, stringx.ToSnake(ctx.GetServiceName())+".go")
	imports := make([]string, 0)
	pbImport := fmt.Sprintf(`"%v"`, ctx.GetPb().Package)
	svcImport := fmt.Sprintf(`"%v"`, ctx.GetSvc().Package)
	configImport := fmt.Sprintf(`"%v"`, ctx.GetConfig().Package)
	imports = append(imports, configImport, pbImport, svcImport)

	var serviceNames []MainServiceTemplateData
	for _, e := range proto.Service {
		var (
			remoteImport string
			serverPkg    string
		)
		if !c.Multiple {
			serverPkg = "server"
			remoteImport = fmt.Sprintf(`"%v"`, ctx.GetServer().Package)
		} else {
			childPkg, err := ctx.GetServer().GetChildPackage(e.Name)
			if err != nil {
				return err
			}

			serverPkg = filepath.Base(childPkg + "Server")
			remoteImport = fmt.Sprintf(`%s "%v"`, serverPkg, childPkg)
		}
		imports = append(imports, remoteImport)
		serviceNames = append(serviceNames, MainServiceTemplateData{
			GRPCService: parser.CamelCase(e.Name),
			Service:     stringx.ToCamel(e.Name),
			ServerPkg:   serverPkg,
			Pkg:         proto.PbPackage,
		})
	}

	text, err := pathx.LoadTemplate(category, mainTemplateFile, mainTemplate)
	if err != nil {
		return err
	}

	return util.With("main").GoFmt(true).Parse(text).SaveTo(map[string]any{
		"serviceName":  stringx.ToSnake(ctx.GetServiceName()),
		"imports":      strings.Join(imports, pathx.NL),
		"pkg":          proto.PbPackage,
		"serviceNames": serviceNames,
	}, fileName, false)
}
