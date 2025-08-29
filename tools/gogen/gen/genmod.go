package gen

import (
	"fmt"
)

type GenMod struct{}

func (g GenMod) Gen(dsl *DSLMeta) (Writer, error) {
	return File{
		Name:    "go.mod",
		Payload: fmt.Sprintf("module %s\n\ngo 1.21\n", dsl.Project),
	}, nil
}
