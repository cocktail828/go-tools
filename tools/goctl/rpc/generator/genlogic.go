package generator

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/collection"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/util"
	"github.com/cocktail828/go-tools/tools/goctl/rpc/parser"
	"github.com/cocktail828/go-tools/z/stringx"
)

const logicFunctionTemplate = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} ({{if .hasReq}}in {{.request}}{{if .stream}},stream {{.streamBody}}{{end}}{{else}}stream {{.streamBody}}{{end}}) ({{if .hasReply}}{{.response}},{{end}} error) {
	// TODO: add your logic here and delete this line
	
	return {{if .hasReply}}&{{.responseType}}{},{{end}} nil
}
`

//go:embed logic.tpl
var logicTemplate string

// GenLogic generates the logic file of the rpc service, which corresponds to the RPC definition items in proto.
func (g *Generator) GenLogic(ctx DirContext, proto parser.Proto, c *ZRpcContext) error {
	if !c.Multiple {
		return g.genLogicInCompatibility(ctx, proto)
	}

	return g.genLogicGroup(ctx, proto)
}

func (g *Generator) genLogicInCompatibility(ctx DirContext, proto parser.Proto) error {
	dir := ctx.GetLogic()
	service := proto.Service[0].Service.Name
	for _, rpc := range proto.Service[0].RPC {
		logicName := fmt.Sprintf("%sLogic", stringx.ToCamel(rpc.Name))
		filename := filepath.Join(dir.Filename, stringx.ToSnake(rpc.Name+"_logic")+".go")
		functions, err := g.genLogicFunction(service, proto.PbPackage, logicName, rpc)
		if err != nil {
			return err
		}

		imports := collection.NewSet()
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetSvc().Package))
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetPb().Package))
		text, err := pathx.LoadTemplate(category, logicTemplateFileFile, logicTemplate)
		if err != nil {
			return err
		}
		err = util.With("logic").GoFmt(true).Parse(text).SaveTo(map[string]any{
			"logicName":   fmt.Sprintf("%sLogic", stringx.ToCamel(rpc.Name)),
			"functions":   functions,
			"packageName": "logic",
			"imports":     strings.Join(imports.KeysStr(), pathx.NL),
		}, filename, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) genLogicGroup(ctx DirContext, proto parser.Proto) error {
	dir := ctx.GetLogic()
	for _, item := range proto.Service {
		serviceName := item.Name
		for _, rpc := range item.RPC {
			var (
				err         error
				filename    string
				logicName   string
				packageName string
			)

			logicName = fmt.Sprintf("%sLogic", stringx.ToCamel(rpc.Name))
			childPkg, err := dir.GetChildPackage(serviceName)
			if err != nil {
				return err
			}

			serviceDir := filepath.Base(childPkg)
			nameJoin := fmt.Sprintf("%s_logic", serviceName)
			packageName = strings.ToLower(stringx.ToCamel(nameJoin))

			filename = filepath.Join(dir.Filename, serviceDir, stringx.ToSnake(rpc.Name+"_logic")+".go")
			functions, err := g.genLogicFunction(serviceName, proto.PbPackage, logicName, rpc)
			if err != nil {
				return err
			}

			imports := collection.NewSet()
			imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetSvc().Package))
			imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetPb().Package))
			text, err := pathx.LoadTemplate(category, logicTemplateFileFile, logicTemplate)
			if err != nil {
				return err
			}

			if err = util.With("logic").GoFmt(true).Parse(text).SaveTo(map[string]any{
				"logicName":   logicName,
				"functions":   functions,
				"packageName": packageName,
				"imports":     strings.Join(imports.KeysStr(), pathx.NL),
			}, filename, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) genLogicFunction(serviceName, goPackage, logicName string, rpc *parser.RPC) (string,
	error) {
	functions := make([]string, 0)
	text, err := pathx.LoadTemplate(category, logicFuncTemplateFileFile, logicFunctionTemplate)
	if err != nil {
		return "", err
	}

	comment := parser.GetComment(rpc.Doc())
	streamServer := fmt.Sprintf("%s.%s_%s%s", goPackage, parser.CamelCase(serviceName),
		parser.CamelCase(rpc.Name), "Server")
	buffer, err := util.With("fun").Parse(text).Execute(map[string]any{
		"logicName":    logicName,
		"method":       parser.CamelCase(rpc.Name),
		"hasReq":       !rpc.StreamsRequest,
		"request":      fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.RequestType)),
		"hasReply":     !rpc.StreamsRequest && !rpc.StreamsReturns,
		"response":     fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"responseType": fmt.Sprintf("%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"stream":       rpc.StreamsRequest || rpc.StreamsReturns,
		"streamBody":   streamServer,
		"hasComment":   len(comment) > 0,
		"comment":      comment,
	})
	if err != nil {
		return "", err
	}

	functions = append(functions, buffer.String())
	return strings.Join(functions, pathx.NL), nil
}
