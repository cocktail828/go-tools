package gogen

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/cocktail828/go-tools/tools/goctl/api/spec"
	"github.com/cocktail828/go-tools/tools/goctl/api/util"
	"github.com/cocktail828/go-tools/tools/goctl/internal/collection"
	"github.com/cocktail828/go-tools/tools/goctl/internal/golang"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/stringx"
)

type fileGenConfig struct {
	rootpath        string // 根路径
	relativepath    string // 相对路径
	filename        string // 文件名
	templateName    string // 模板名
	category        string
	templateFile    string
	builtinTemplate string
	data            any
}

func genFile(c fileGenConfig) error {
	fp, created, err := util.ShouldCreateFile(c.rootpath, c.relativepath, c.filename)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	var text string
	if len(c.category) == 0 || len(c.templateFile) == 0 {
		text = c.builtinTemplate
	} else {
		text, err = pathx.LoadTemplate(c.category, c.templateFile, c.builtinTemplate)
		if err != nil {
			return err
		}
	}

	t := template.Must(template.New(c.templateName).Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, c.data)
	if err != nil {
		return err
	}

	code := golang.FormatCode(buffer.String())
	_, err = fp.WriteString(code)
	return err
}

func writeProperty(writer io.Writer, name, tag, comment string, tp spec.Type, indent int) error {
	util.WriteIndent(writer, indent)
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

func getMiddleware(api *spec.ApiSpec) []string {
	result := collection.NewSet()
	for _, g := range api.Service.Groups {
		middleware := g.GetAnnotation("middleware")
		if len(middleware) > 0 {
			for _, item := range strings.Split(middleware, ",") {
				result.Add(strings.TrimSpace(item))
			}
		}
	}

	return result.KeysStr()
}

func responseGoTypeName(r spec.Route, pkg ...string) string {
	if r.ResponseType == nil {
		return ""
	}

	resp := golangExpr(r.ResponseType, pkg...)
	switch r.ResponseType.(type) {
	case spec.DefineStruct:
		if !strings.HasPrefix(resp, "*") {
			return "*" + resp
		}
	}

	return resp
}

func requestGoTypeName(r spec.Route, pkg ...string) string {
	if r.RequestType == nil {
		return ""
	}

	return golangExpr(r.RequestType, pkg...)
}

func golangExpr(ty spec.Type, pkg ...string) string {
	switch v := ty.(type) {
	case spec.PrimitiveType:
		return v.RawName
	case spec.DefineStruct:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}

		if len(pkg) == 0 {
			return v.RawName
		}

		return fmt.Sprintf("%s.%s", pkg[0], stringx.Title(v.RawName))
	case spec.ArrayType:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}

		if len(pkg) == 0 {
			return v.RawName
		}

		return fmt.Sprintf("[]%s", golangExpr(v.Value, pkg...))
	case spec.MapType:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}

		if len(pkg) == 0 {
			return v.RawName
		}

		return fmt.Sprintf("map[%s]%s", v.Key, golangExpr(v.Value, pkg...))
	case spec.PointerType:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}

		if len(pkg) == 0 {
			return v.RawName
		}

		return fmt.Sprintf("*%s", golangExpr(v.Type, pkg...))
	case spec.InterfaceType:
		return v.RawName
	}

	return ""
}

func getDoc(doc string) string {
	if len(doc) == 0 {
		return ""
	}

	return "// " + strings.Trim(doc, "\"")
}
