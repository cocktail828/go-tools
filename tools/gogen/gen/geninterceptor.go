package gen

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

var (
	//go:embed geninterceptor.tpl
	geninterceptorTpl string
)

type GenInterceptor struct{}

func (g GenInterceptor) Gen(dsl *DSLMeta) (Writer, error) {
	ws := MultiFile{}
	interceptorSet := map[string]struct{}{}

	tpl, err := template.New("interceptor").Parse(geninterceptorTpl)
	if err != nil {
		return nil, err
	}

	genViaTpl := func(incps ...string) error {
		for _, ic := range incps {
			if _, ok := interceptorSet[ic]; ok {
				return errors.Errorf("interceptor[%v] has already been defined", ic)
			}
			interceptorSet[ic] = struct{}{}

			sb := strings.Builder{}
			if err := tpl.Execute(&sb, map[string]any{
				"name": ic,
			}); err != nil {
				return err
			}

			ws = append(ws, File{
				SubDir:  "interceptor",
				Name:    strings.ToLower(ic) + ".go",
				Payload: sb.String(),
			})
		}

		return nil
	}

	for _, svc := range dsl.Services {
		if err := genViaTpl(svc.Interceptors...); err != nil {
			return nil, err
		}

		for _, grp := range svc.Groups {
			if err := genViaTpl(grp.Interceptors...); err != nil {
				return nil, err
			}
		}
	}

	return ws, nil
}
