package gogen

import (
	_ "embed"
	"fmt"
	"io"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/stringx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/util"
	"github.com/cocktail828/go-tools/tools/goctl/internal/version"
	"github.com/pkg/errors"
)

//go:embed model.tpl
var typesTemplate string

// BuildModel gen types to string
func BuildModel(types []spec.Type) (string, error) {
	var builder strings.Builder
	first := true
	for _, tp := range types {
		if first {
			first = false
		} else {
			builder.WriteString("\n\n")
		}
		if err := writeType(&builder, tp); err != nil {
			return "", errors.Wrapf(err, "Type "+tp.Name()+" generate error")
		}
	}

	return builder.String(), nil
}

func genModel(dir string, api *spec.ApiSpec) error {
	val, err := BuildModel(api.Types)
	if err != nil {
		return err
	}

	return genFile(fileGenConfig{
		rootpath:         dir,
		relativepath:     typesDir,
		filename:         "model.go",
		templateName:     "typesTemplate",
		category:         category,
		templateFileName: typesTemplateFile,
		builtinTemplate:  typesTemplate,
		data: map[string]any{
			"types":        val,
			"containsTime": false,
			"version":      version.BuildVersion,
		},
	})
}

func writeType(writer io.Writer, tp spec.Type) error {
	structType, ok := tp.(spec.DefineStruct)
	if !ok {
		return fmt.Errorf("unspport struct type: %s", tp.Name())
	}

	_, err := fmt.Fprintf(writer, "type %s struct {\n", util.Title(tp.Name()))
	if err != nil {
		return err
	}

	if err := writeMember(writer, structType.Members); err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "}")
	return err
}

func writeMember(writer io.Writer, members []spec.Member) error {
	for _, member := range members {
		if member.IsInline {
			if _, err := fmt.Fprintf(writer, "%s\n", stringx.Title(member.Type.Name())); err != nil {
				return err
			}
			continue
		}

		if err := writeProperty(writer, member.Name, member.Tag, member.GetComment(), member.Type, 1); err != nil {
			return err
		}
	}
	return nil
}
