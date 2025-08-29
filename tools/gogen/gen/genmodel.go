package gen

import (
	_ "embed"
	"strings"
	"text/template"
)

var (
	//go:embed genmodel.tpl
	genmodelTpl string
)

type GenModel struct{}

func (g GenModel) Gen(dsl *DSLMeta) (Writer, error) {
	set := map[string]struct{}{}

	tpl, err := template.New("model").Parse(genmodelTpl)
	if err != nil {
		return nil, err
	}

	payloads := []string{"package model\n"}
	for _, st := range dsl.Structs {
		if _, ok := set[st.Name]; ok {
			continue
		}
		set[st.Name] = struct{}{}

		sb := strings.Builder{}
		if err := tpl.Execute(&sb, st); err != nil {
		}
		payloads = append(payloads, sb.String())
	}

	return File{
		SubDir:  "model",
		Name:    "model.go",
		Payload: strings.Join(payloads, "\n"),
	}, nil
}
