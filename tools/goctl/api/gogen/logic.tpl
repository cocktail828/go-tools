package {{.pkgName}}

import (
	"context"
	
	"github.com/cocktail828/go-tools/xlog"
	{{.imports}}
)

type {{.logic}} struct {
	xlog.Logger
	ctx    context.Context
}

{{if .hasDoc}}{{.doc}}{{end}}
func New{{.logic}}(ctx context.Context, log xlog.Logger) *{{.logic}} {
	return &{{.logic}}{
		Logger: log,
		ctx:    ctx,
	}
}

func (l *{{.logic}}) {{.function}}({{.request}}) {{.responseType}} {
	// TODO: add your logic here and delete this line

	{{.returnString}}
}
