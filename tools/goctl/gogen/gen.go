package gogen

import (
	"bytes"
	"log"
	"os"
	"path"
	"sync"
	"text/template"

	"github.com/cocktail828/go-tools/tools/goctl/internal/golang"
	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/parser"
	"github.com/cocktail828/go-tools/tools/goctl/internal/parser/spec"
)

const (
	groupProperty = "group"
)

var (
	templ = Template{}
)

type Render interface {
	Render(string) error
}

type ErrRender struct{ Err error }

func (r ErrRender) Render(string) error { return r.Err }

type MultiRender []Render

func (r MultiRender) Render(home string) error {
	for _, f := range r {
		if err := f.Render(home); err != nil {
			return err
		}
	}
	return nil
}

type NopRender struct {
	rootpath     string // 根路径
	relativepath string // 相对路径
	filename     string // 文件名
	data         []byte
}

func (r NopRender) Render(home string) error {
	payload := golang.FormatCode(string(r.data))

	dir := path.Join(r.rootpath, r.relativepath)
	if dir != "" {
		// MkdirAll is idempotent
		if err := os.MkdirAll(path.Join(r.rootpath, r.relativepath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}
	}

	return os.WriteFile(path.Join(r.rootpath, r.relativepath, r.filename), []byte(payload), os.ModePerm)
}

type FileRender struct {
	rootpath         string // 根路径
	relativepath     string // 相对路径
	filename         string // 文件名
	templateFileName string // 模板文件名
	data             any    // 模板参数
}

func (r FileRender) Render(home string) error {
	body, err := templ.Load(home, r.templateFileName)
	if err != nil {
		return err
	}

	t := template.Must(template.New(r.templateFileName).Parse(string(body)))
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, r.data); err != nil {
		return err
	}

	payload := golang.FormatCode(buffer.String())

	dir := path.Join(r.rootpath, r.relativepath)
	if dir != "" {
		// MkdirAll is idempotent
		if err := os.MkdirAll(path.Join(r.rootpath, r.relativepath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}
	}

	return os.WriteFile(path.Join(r.rootpath, r.relativepath, r.filename), []byte(payload), os.ModePerm)
}

type Type string

const (
	TypeVars       = "vars"
	TypeMakefile   = "makefile"
	TypeMain       = "main"
	TypeService    = "service"
	TypeMiddleware = "middleware"
	TypeRoute      = "route"
	TypeModel      = "model"
)

type Export struct {
	Vars    []string
	Structs []string
	Funcs   []string
}

type FileMeta struct {
	RootPath string
	Mod      string
	Lookup   func(ty Type) Generater
}

type Generater interface {
	Init(*spec.ApiSpec) error
	PkgName() string      // 包名
	RelativePath() string // 相对go mod 的子路径
	Export() Export       // 导出信息
	Gen(FileMeta) Render
}

var (
	generaterMu  = sync.RWMutex{}
	generaterMap = map[Type]Generater{}
)

func Register(tp Type, src Generater) {
	generaterMu.Lock()
	defer generaterMu.Unlock()
	if _, ok := generaterMap[tp]; ok {
		log.Fatalf("generater %q already exist", tp)
	}
	generaterMap[tp] = src
}

// genGoProject gen go project files with api File
func genGoProject(modname string, dir, home, apiFile string) {
	api, err := parser.Parse(apiFile, "")
	if err != nil {
		log.Fatal(err)
	}

	if err := api.Validate(); err != nil {
		log.Fatal(err)
	}

	for _, g := range generaterMap {
		if err := g.Init(api); err != nil {
			log.Fatal(err)
		}
	}

	for _, g := range generaterMap {
		r := g.Gen(FileMeta{
			RootPath: dir,
			Mod:      modname,
			Lookup: func(ty Type) Generater {
				generaterMu.RLock()
				defer generaterMu.RUnlock()
				src, ok := generaterMap[ty]
				if !ok {
					log.Fatalf("%q is required, but not found", ty)
				}
				return src
			},
		})

		if err := r.Render(home); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Congratulations! Your project has been successfully generated.")
	log.Println("To initialize your project, run:")
	log.Println("    make init")
	log.Println()
	log.Println("Enjoy your development journey! 🚀")
}
