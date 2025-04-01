package gogen

import (
	_ "embed"
	"fmt"
	"io"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/version"
	"github.com/cocktail828/go-tools/z/stringx"
	"github.com/pkg/errors"
)

func init() { Register(TypeModel, &modelGenerater{}) }

type modelGenerater struct {
	*spec.ApiSpec
}

func (g *modelGenerater) PkgName() string              { return "model" }
func (g *modelGenerater) RelativePath() string         { return "model" }
func (g *modelGenerater) TemplateFile() string         { return "model.tpl" }
func (g *modelGenerater) Init(api *spec.ApiSpec) error { g.ApiSpec = api; return nil }
func (g *modelGenerater) Export() Export               { return Export{} }
func (g *modelGenerater) Gen(fm FileMeta) Render {
	var builder strings.Builder
	first := true
	for _, tp := range g.Types {
		if first {
			first = false
		} else {
			builder.WriteString("\n\n")
		}
		if err := writeType(&builder, tp); err != nil {
			return ErrRender{errors.Wrapf(err, "Type "+tp.Name()+" generate error")}
		}
	}

	return FileRender{
		rootpath:         fm.RootPath,
		relativepath:     g.RelativePath(),
		filename:         "model.go",
		templateFileName: g.TemplateFile(),
		data: map[string]any{
			"types":        builder.String(),
			"containsTime": false,
			"version":      version.BuildVersion,
		},
	}
}

func writeType(writer io.Writer, tp spec.Type) error {
	structType, ok := tp.(spec.DefineStruct)
	if !ok {
		return errors.Errorf("unspport struct type: %s", tp.Name())
	}

	_, err := fmt.Fprintf(writer, "type %s struct {\n", stringx.Title(tp.Name()))
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

func writeProperty(writer io.Writer, name, tag, comment string, tp spec.Type, indent int) error {
	fmt.Fprintln(writer, strings.Repeat("\t", indent))
	var (
		err            error
		isNestedStruct bool
	)
	structType, ok := tp.(spec.NestedStruct)
	if ok {
		isNestedStruct = true
	}
	if len(comment) > 0 {
		comment = strings.TrimPrefix(comment, "//")
		comment = "//" + comment
	}

	if isNestedStruct {
		_, err = fmt.Fprintf(writer, "%s struct {\n", stringx.Title(name))
		if err != nil {
			return err
		}

		if err := writeMember(writer, structType.Members); err != nil {
			return err
		}

		_, err := fmt.Fprintf(writer, "} %s", tag)
		if err != nil {
			return err
		}

		if len(comment) > 0 {
			_, err = fmt.Fprintf(writer, " %s", comment)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprint(writer, "\n")
		if err != nil {
			return err
		}
	} else {
		if len(comment) > 0 {
			_, err = fmt.Fprintf(writer, "%s %s %s %s\n", stringx.Title(name), tp.Name(), tag, comment)
			if err != nil {
				return err
			}
		} else {
			_, err = fmt.Fprintf(writer, "%s %s %s\n", stringx.Title(name), tp.Name(), tag)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
