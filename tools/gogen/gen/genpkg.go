package gen

import (
	_ "embed"
	"strings"
	"text/template"
)

var (
	//go:embed genpkg.tpl
	pkgTpl string
)

type GenPkg struct{}

func (g GenPkg) Gen(dsl *DSLMeta) (Writer, error) {
	tpl, err := template.New("pkg").Parse(pkgTpl)
	if err != nil {
		return nil, err
	}

	ws := MultiFile{}
	for _, svc := range dsl.Services {
		sb := strings.Builder{}

		if err := tpl.Execute(&sb, map[string]any{
			"has_interceptor": svc.HasInterceptor,
			"project":         dsl.Project,
			"service":         svc,
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
