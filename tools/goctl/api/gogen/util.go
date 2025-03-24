package gogen

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/cocktail828/go-tools/tools/goctl/api/util"
	"github.com/cocktail828/go-tools/tools/goctl/internal/collection"
	"github.com/cocktail828/go-tools/tools/goctl/internal/golang"
	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/stringx"
)

type fileGenConfig struct {
	rootpath         string // 根路径
	relativepath     string // 相对路径
	filename         string // 文件名
	templateName     string // 模板名
	templateFileName string // 模板文件名
	category         string // 自定义模板路径
	builtinTemplate  string // 内置模板
	data             any    // 模板参数
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
	if len(c.category) == 0 || len(c.templateFileName) == 0 {
		text = c.builtinTemplate
	} else {
		text, err = pathx.LoadTemplate(c.category, c.templateFileName, c.builtinTemplate)
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

func getDoc(doc string) string {
	if len(doc) == 0 {
		return ""
	}

	return "// " + strings.Trim(doc, "\"")
}
