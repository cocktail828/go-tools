package gen

import (
	"github.com/cocktail828/go-tools/tools/gogen/ast"
)

type Generater interface {
	Gen(svc *DSLMeta) (Writer, error)
}

func Generate(root string, dsl *ast.DSL) error {
	gs := []Generater{GenMod{}, GenPkg{}, GenInterceptor{}, GenHandler{}, GenModel{}, GenConfig{}}
	meta := DSLMeta{
		Syntax:  dsl.Syntax,
		Project: dsl.Project,
		Structs: dsl.Structs,
	}

	for _, svc := range dsl.Services {
		meta.Services = append(meta.Services, serviceAst2Meta(svc))
	}

	for _, g := range gs {
		wr, err := g.Gen(&meta)
		if err != nil {
			return err
		}

		err = wr.Write("xxx")
		if err != nil {
			return err
		}
	}
	return nil
}
