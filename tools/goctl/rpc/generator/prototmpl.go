package generator

import (
	_ "embed"
	"path/filepath"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/util"
	"github.com/cocktail828/go-tools/z/stringx"
)

//go:embed rpc.tpl
var rpcTemplateText string

// ProtoTmpl returns a sample of a proto file
func ProtoTmpl(out string) error {
	protoFilename := filepath.Base(out)
	serviceName := strings.TrimSuffix(protoFilename, filepath.Ext(protoFilename))
	text, err := pathx.LoadTemplate(category, rpcTemplateFile, rpcTemplateText)
	if err != nil {
		return err
	}

	dir := filepath.Dir(out)
	err = pathx.MkdirIfNotExist(dir)
	if err != nil {
		return err
	}

	err = util.With("t").Parse(text).SaveTo(map[string]string{
		"package":     stringx.Untitle(serviceName),
		"serviceName": stringx.Title(serviceName),
	}, out, false)
	return err
}
