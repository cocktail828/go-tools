package gen

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/cocktail828/go-tools/tools/gogen/ast"
)

var (
	//go:embed genmodel.tpl
	genmodelTpl string
)

type GenModel struct{}

func (g GenModel) Gen(dsl *ast.DSL) (Writer, error) {
	set := map[string]struct{}{}

	tpl, err := template.New("model").Parse(genmodelTpl)
	if err != nil {
		return nil, err
	}

	gen := func(name string) string {
		name = strings.Title(name)
		if _, ok := set[name]; ok {
			return ""
		}
		set[name] = struct{}{}

		sb := strings.Builder{}
		if err := tpl.Execute(&sb, map[string]any{
			"name": name,
		}); err != nil {
		}
		return sb.String()
	}

	payloads := []string{"package model\n"}
	for _, svc := range dsl.Services {
		for _, grp := range svc.Groups {
			for _, rt := range grp.Routes {
				if rt.Request != "" {
					payloads = append(payloads, gen(rt.Request))
				}
				if rt.Response != "" {
					payloads = append(payloads, gen(rt.Response))
				}
			}
		}
	}

	return File{
		Path:    "model",
		Name:    "model.go",
		Payload: strings.Join(payloads, "\n"),
	}, nil
}
