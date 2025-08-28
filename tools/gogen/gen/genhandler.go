package gen

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/cocktail828/go-tools/tools/gogen/ast"
)

var (
	//go:embed genhandler.tpl
	genhandlerTpl string
)

type GenHandler struct{}

func (g GenHandler) Gen(dsl *ast.DSL) (Writer, error) {
	ws := MultiFile{}
	interceptorSet := map[string]struct{}{}

	tpl, err := template.New("handler").Parse(genhandlerTpl)
	if err != nil {
		return nil, err
	}

	genViaTpl := func(name string) error {
		if _, ok := interceptorSet[name]; ok {
			return errors.Errorf("handler [%v] has already been defined", name)
		}

		interceptorSet[name] = struct{}{}

		sb := strings.Builder{}
		if err := tpl.Execute(&sb, map[string]any{
			"name": name,
		}); err != nil {
			return err
		}

		ws = append(ws, File{
			Path:    "handler",
			Name:    strings.ToLower(name) + "_handler.go",
			Payload: sb.String(),
		})

		return nil
	}

	for _, svc := range dsl.Services {
		for _, ic := range svc.Interceptors {
			if err := genViaTpl(ic); err != nil {
				return nil, err
			}
		}

		for _, grp := range svc.Groups {
			for _, ic := range grp.Interceptors {
				if err := genViaTpl(ic); err != nil {
					return nil, err
				}
			}
		}
	}

	return ws, nil
}
