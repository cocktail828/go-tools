package gen

import (
	"os"

	"github.com/cocktail828/go-tools/tools/gogen/ast"
)

type Generater interface {
	Gen(dsl *ast.DSL) (Writer, error)
}

func Generate(root string, dsl *ast.DSL) error {
	gs := []Generater{GenMod{}, GenPkg{}, GenInterceptor{}, GenHandler{}, GenModel{}}

	os.MkdirAll(root, 0755)
	for _, g := range gs {
		wr, err := g.Gen(dsl)
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
