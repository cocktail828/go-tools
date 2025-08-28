package gen

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/cocktail828/go-tools/tools/gogen/ast"
)

var (
	//go:embed genpkg.tpl
	pkgTpl string
)

type GenPkg struct{}

func (g GenPkg) Gen(dsl *ast.DSL) (Writer, error) {
	tpl, err := template.New("pkg").Parse(pkgTpl)
	if err != nil {
		return nil, err
	}

	ws := MultiFile{}
	for _, svc := range dsl.Services {
		sb := strings.Builder{}
		if err := tpl.Execute(&sb, map[string]any{
			"interceptors": svc.Interceptors,
			"has_interceptor": func() bool {
				if len(svc.Interceptors) != 0 {
					return true
				}

				for _, grp := range svc.Groups {
					if len(grp.Interceptors) != 0 {
						return true
					}
				}

				return false
			},
			"project": dsl.Project,
			"service": svc,
		}); err != nil {
			return nil, err
		}

		ws = append(ws, File{
			Name:    "main.go",
			Payload: sb.String(),
		})
	}

	return ws, nil
}
