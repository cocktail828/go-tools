package gen

import (
	_ "embed"
	"strings"
	"text/template"
)

var (
	//go:embed genhandler.tpl
	genhandlerTpl string
)

type GenHandler struct{}

func (g GenHandler) Gen(dsl *DSLMeta) (Writer, error) {
	ws := MultiFile{}

	tpl, err := template.New("handler").Parse(genhandlerTpl)
	if err != nil {
		return nil, err
	}

	for _, svc := range dsl.Services {
		for _, grp := range svc.Groups {
			for _, rt := range grp.Routes {
				sb := strings.Builder{}
				if err := tpl.Execute(&sb, map[string]any{
					"project": dsl.Project,
					"route":   rt,
				}); err != nil {
					return nil, err
				}

				ws = append(ws, File{
					SubDir:  "handler",
					Name:    strings.ToLower(rt.HandlerName) + ".go",
					Payload: sb.String(),
				})
			}
		}
	}

	return ws, nil
}
