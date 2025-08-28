package gen

import (
	"fmt"

	"github.com/cocktail828/go-tools/tools/gogen/ast"
)

type GenMod struct{}

func (g GenMod) Gen(dsl *ast.DSL) (Writer, error) {
	return File{
		Name:    "go.mod",
		Payload: fmt.Sprintf("module %s\n\ngo 1.21\n", dsl.Project),
	}, nil
}
